package search

import (
	"bufio"
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// SearchResult represents a search provider result
type SearchResult struct {
	FilePath   string  `json:"file_path"`
	LineNumber int     `json:"line_number,omitempty"`
	Content    string  `json:"content"`
	Score      float64 `json:"score,omitempty"`
	MatchType  string  `json:"match_type"` // "semantic", "keyword", "regex"
}

// SearchOptions configures search behavior
type SearchOptions struct {
	Limit       int
	Offset      int
	CaseSensitive bool
	Regex       bool
	FilePattern string // glob pattern for file filtering
	Directory   string // search within directory
}

// Provider defines the interface for search providers
type Provider interface {
	// Search performs a search and returns results
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

	// Available returns true if the provider is available
	Available() bool

	// Name returns the provider name
	Name() string
}

// VecgrepProvider uses vecgrep for semantic search
type VecgrepProvider struct {
	projectPath string
}

// NewVecgrepProvider creates a new vecgrep provider
func NewVecgrepProvider(projectPath string) *VecgrepProvider {
	return &VecgrepProvider{projectPath: projectPath}
}

// Available checks if vecgrep is installed and the project is indexed
func (v *VecgrepProvider) Available() bool {
	_, err := exec.LookPath("vecgrep")
	return err == nil
}

// Name returns the provider name
func (v *VecgrepProvider) Name() string {
	return "vecgrep"
}

// Search performs a semantic search using vecgrep
func (v *VecgrepProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if !v.Available() {
		return nil, ErrProviderUnavailable
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	// Build vecgrep command
	args := []string{"search", query, "--limit", strconv.Itoa(limit)}
	if opts.Directory != "" {
		args = append(args, "--directory", opts.Directory)
	}
	if opts.FilePattern != "" {
		args = append(args, "--file-pattern", opts.FilePattern)
	}

	cmd := exec.CommandContext(ctx, "vecgrep", args...)
	cmd.Dir = v.projectPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseVecgrepOutput(output), nil
}

// vecgrepResult represents a single result from vecgrep JSON output
type vecgrepResult struct {
	File       string  `json:"file"`
	FilePath   string  `json:"file_path"`
	Line       int     `json:"line"`
	StartLine  int     `json:"start_line"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	Similarity float64 `json:"similarity"`
}

// parseVecgrepOutput parses vecgrep output (JSON or line-based)
func parseVecgrepOutput(output []byte) []SearchResult {
	if len(output) == 0 {
		return nil
	}

	// Try JSON array first
	var jsonResults []vecgrepResult
	if err := json.Unmarshal(output, &jsonResults); err == nil {
		results := make([]SearchResult, 0, len(jsonResults))
		for _, r := range jsonResults {
			filePath := r.FilePath
			if filePath == "" {
				filePath = r.File
			}
			lineNum := r.Line
			if lineNum == 0 {
				lineNum = r.StartLine
			}
			score := r.Score
			if score == 0 {
				score = r.Similarity
			}
			results = append(results, SearchResult{
				FilePath:   filePath,
				LineNumber: lineNum,
				Content:    r.Content,
				Score:      score,
				MatchType:  "semantic",
			})
		}
		return results
	}

	// Try JSONL (one JSON object per line)
	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var r vecgrepResult
		if err := json.Unmarshal([]byte(line), &r); err == nil {
			filePath := r.FilePath
			if filePath == "" {
				filePath = r.File
			}
			lineNum := r.Line
			if lineNum == 0 {
				lineNum = r.StartLine
			}
			score := r.Score
			if score == 0 {
				score = r.Similarity
			}
			results = append(results, SearchResult{
				FilePath:   filePath,
				LineNumber: lineNum,
				Content:    r.Content,
				Score:      score,
				MatchType:  "semantic",
			})
		}
	}

	return results
}

// FallbackProvider uses grep/ripgrep for text search
type FallbackProvider struct {
	projectPath string
}

// NewFallbackProvider creates a new fallback text search provider
func NewFallbackProvider(projectPath string) *FallbackProvider {
	return &FallbackProvider{projectPath: projectPath}
}

// Available always returns true as fallback is always available
func (f *FallbackProvider) Available() bool {
	return true
}

// Name returns the provider name
func (f *FallbackProvider) Name() string {
	return "text"
}

// Search performs a text search using grep
func (f *FallbackProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	// Try ripgrep first, then fall back to grep
	var cmd *exec.Cmd

	if _, err := exec.LookPath("rg"); err == nil {
		args := []string{query, "--json"}
		if !opts.CaseSensitive {
			args = append(args, "-i")
		}
		if opts.FilePattern != "" {
			args = append(args, "-g", opts.FilePattern)
		}
		args = append(args, "-m", strconv.Itoa(limit))
		cmd = exec.CommandContext(ctx, "rg", args...)
	} else {
		args := []string{"-r", "-n", query}
		if !opts.CaseSensitive {
			args = append(args, "-i")
		}
		if opts.FilePattern != "" {
			args = append(args, "--include="+opts.FilePattern)
		}
		cmd = exec.CommandContext(ctx, "grep", args...)
	}

	cmd.Dir = f.projectPath
	if opts.Directory != "" {
		cmd.Dir = opts.Directory
	}

	output, _ := cmd.Output() // grep returns non-zero if no matches
	return parseGrepOutput(output, limit), nil
}

// parseGrepOutput parses grep/ripgrep output in file:line:content format
func parseGrepOutput(output []byte, limit int) []SearchResult {
	if len(output) == 0 {
		return nil
	}

	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() && len(results) < limit {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Try JSON format first (ripgrep --json)
		var rgJSON struct {
			Type string `json:"type"`
			Data struct {
				Path struct {
					Text string `json:"text"`
				} `json:"path"`
				LineNumber int `json:"line_number"`
				Lines      struct {
					Text string `json:"text"`
				} `json:"lines"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(line), &rgJSON); err == nil {
			if rgJSON.Type == "match" {
				results = append(results, SearchResult{
					FilePath:   rgJSON.Data.Path.Text,
					LineNumber: rgJSON.Data.LineNumber,
					Content:    strings.TrimRight(rgJSON.Data.Lines.Text, "\n"),
					MatchType:  "keyword",
				})
			}
			continue
		}

		// Fall back to standard grep format: file:line:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			lineNum := 0
			if n, err := strconv.Atoi(parts[1]); err == nil {
				lineNum = n
			}
			results = append(results, SearchResult{
				FilePath:   parts[0],
				LineNumber: lineNum,
				Content:    parts[2],
				MatchType:  "keyword",
			})
		} else if len(parts) == 2 {
			// file:content (no line number)
			results = append(results, SearchResult{
				FilePath:  parts[0],
				Content:   parts[1],
				MatchType: "keyword",
			})
		}
	}

	return results
}

// Errors
type searchError string

func (e searchError) Error() string { return string(e) }

const (
	ErrProviderUnavailable = searchError("search provider unavailable")
)

// MultiProvider tries multiple providers in order
type MultiProvider struct {
	providers []Provider
}

// NewMultiProvider creates a provider that tries multiple providers
func NewMultiProvider(providers ...Provider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

// Available returns true if any provider is available
func (m *MultiProvider) Available() bool {
	for _, p := range m.providers {
		if p.Available() {
			return true
		}
	}
	return false
}

// Name returns the name of the first available provider
func (m *MultiProvider) Name() string {
	for _, p := range m.providers {
		if p.Available() {
			return p.Name()
		}
	}
	return "none"
}

// Search uses the first available provider
func (m *MultiProvider) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	for _, p := range m.providers {
		if p.Available() {
			return p.Search(ctx, query, opts)
		}
	}
	return nil, ErrProviderUnavailable
}
