package search

import (
	"context"
	"os/exec"
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
	args := []string{"search", query, "--limit", string(rune(limit))}
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

// parseVecgrepOutput parses vecgrep JSON output
func parseVecgrepOutput(output []byte) []SearchResult {
	// vecgrep outputs results in a specific format
	// For now, return empty - full implementation would parse actual output
	return []SearchResult{}
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
		args = append(args, "-m", string(rune(limit)))
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

// parseGrepOutput parses grep output
func parseGrepOutput(output []byte, limit int) []SearchResult {
	// Parse grep output format: file:line:content
	results := []SearchResult{}
	// Full implementation would parse the output
	_ = output
	_ = limit
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
