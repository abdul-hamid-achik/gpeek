package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
)

// CommitDocument represents a commit formatted for embedding
type CommitDocument struct {
	Hash     string   `json:"hash"`
	Message  string   `json:"message"`
	Author   string   `json:"author"`
	Files    []string `json:"files"`
	FullText string   `json:"full_text"`
}

// SearchResult represents a semantic search result
type SearchResult struct {
	Hash       string  `json:"hash"`
	Message    string  `json:"message"`
	Author     string  `json:"author"`
	Similarity float64 `json:"similarity"`
	MatchedOn  string  `json:"matched_on"`
}

// VecgrepIntegration handles vecgrep operations
type VecgrepIntegration struct {
	repoPath  string
	indexPath string
}

// NewVecgrepIntegration creates a new vecgrep integration
func NewVecgrepIntegration(repoPath string) *VecgrepIntegration {
	indexPath := filepath.Join(repoPath, ".gpeek")
	return &VecgrepIntegration{
		repoPath:  repoPath,
		indexPath: indexPath,
	}
}

// EnsureIndexDir creates the index directory if it doesn't exist
func (v *VecgrepIntegration) EnsureIndexDir() error {
	if err := os.MkdirAll(v.indexPath, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Add .gpeek to .gitignore if not already present
	gitignorePath := filepath.Join(v.repoPath, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err == nil {
		if !strings.Contains(string(content), ".gpeek") {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				_, _ = f.WriteString("\n.gpeek/\n")
				_ = f.Close()
			}
		}
	}

	return nil
}

// IndexCommits indexes commits for semantic search
func (v *VecgrepIntegration) IndexCommits(repo *git.Repository, limit int) error {
	if err := v.EnsureIndexDir(); err != nil {
		return err
	}

	commits, err := repo.Log(limit)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	// Create a text file with commit information for each commit
	commitsDir := filepath.Join(v.indexPath, "commits")
	if err := os.MkdirAll(commitsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commits directory: %w", err)
	}

	for _, c := range commits {
		// Get files changed in this commit
		diff, _ := repo.CommitDiff(c.Hash)
		files := extractFilesFromDiff(diff)

		// Create document text that includes commit message and files
		docText := fmt.Sprintf("Commit: %s\nAuthor: %s\nMessage: %s\nFiles: %s",
			c.Hash[:7],
			c.Author,
			c.Message,
			strings.Join(files, ", "))

		// Write to file
		docPath := filepath.Join(commitsDir, c.Hash[:7]+".txt")
		if err := os.WriteFile(docPath, []byte(docText), 0644); err != nil {
			return fmt.Errorf("failed to write commit document: %w", err)
		}
	}

	// Index with vecgrep
	cmd := exec.Command("vecgrep", "index", commitsDir)
	cmd.Dir = v.indexPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vecgrep index failed: %s", string(output))
	}

	return nil
}

// SearchCommits performs semantic search on indexed commits
func (v *VecgrepIntegration) SearchCommits(query string, limit int) ([]SearchResult, error) {
	// Use vecgrep to search
	cmd := exec.Command("vecgrep", "search", query, "--limit", fmt.Sprintf("%d", limit), "--format", "json")
	cmd.Dir = v.indexPath
	output, err := cmd.Output()
	if err != nil {
		// If vecgrep is not available, fall back to simple text search
		return v.fallbackSearch(query, limit)
	}

	var results []SearchResult
	if err := json.Unmarshal(output, &results); err != nil {
		// Try parsing line by line
		return v.parseVecgrepOutput(output, limit)
	}

	return results, nil
}

// fallbackSearch performs a simple text search when vecgrep is not available
func (v *VecgrepIntegration) fallbackSearch(query string, limit int) ([]SearchResult, error) {
	commitsDir := filepath.Join(v.indexPath, "commits")

	files, err := os.ReadDir(commitsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read commits directory: %w", err)
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(commitsDir, f.Name()))
		if err != nil {
			continue
		}

		contentLower := strings.ToLower(string(content))
		if strings.Contains(contentLower, queryLower) {
			// Parse the content to extract fields
			hash := strings.TrimSuffix(f.Name(), ".txt")
			lines := strings.Split(string(content), "\n")

			var message, author string
			for _, line := range lines {
				if strings.HasPrefix(line, "Message: ") {
					message = strings.TrimPrefix(line, "Message: ")
				}
				if strings.HasPrefix(line, "Author: ") {
					author = strings.TrimPrefix(line, "Author: ")
				}
			}

			results = append(results, SearchResult{
				Hash:       hash,
				Message:    message,
				Author:     author,
				Similarity: 0.5, // Basic match
				MatchedOn:  "text",
			})

			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// parseVecgrepOutput parses vecgrep output when JSON parsing fails
func (v *VecgrepIntegration) parseVecgrepOutput(output []byte, limit int) ([]SearchResult, error) {
	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() && len(results) < limit {
		line := scanner.Text()
		// Parse vecgrep output format (usually file path with similarity score)
		// Format varies by version, so we'll extract what we can
		if strings.Contains(line, ".txt") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				filename := filepath.Base(parts[0])
				hash := strings.TrimSuffix(filename, ".txt")

				similarity := 0.5
				if len(parts) >= 2 {
					_, _ = fmt.Sscanf(parts[1], "%f", &similarity)
				}

				results = append(results, SearchResult{
					Hash:       hash,
					Similarity: similarity,
					MatchedOn:  "semantic",
				})
			}
		}
	}

	return results, nil
}

// FindSimilarToStaged finds commits similar to current staged changes
func (v *VecgrepIntegration) FindSimilarToStaged(repo *git.Repository, limit int) ([]SearchResult, error) {
	status, err := repo.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	if len(status.Staged) == 0 {
		return nil, fmt.Errorf("no staged changes")
	}

	// Build a query from staged files and their changes
	var queryParts []string
	for _, f := range status.Staged {
		queryParts = append(queryParts, f.Path)
		// Get diff for context
		diff, _ := repo.FileDiff(f.Path, true)
		if diff != "" {
			// Extract just the added/removed lines for query
			lines := strings.Split(diff, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					queryParts = append(queryParts, strings.TrimPrefix(line, "+"))
				}
			}
		}
	}

	// Limit query size
	query := strings.Join(queryParts, " ")
	if len(query) > 500 {
		query = query[:500]
	}

	return v.SearchCommits(query, limit)
}

// FindSimilarToFile finds commits that touched similar code
func (v *VecgrepIntegration) FindSimilarToFile(repo *git.Repository, filePath string, limit int) ([]SearchResult, error) {
	// Read file content
	content, err := os.ReadFile(filepath.Join(repo.Path(), filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Use file path and first few lines as query
	lines := strings.Split(string(content), "\n")
	queryLines := lines
	if len(queryLines) > 20 {
		queryLines = queryLines[:20]
	}

	query := filePath + " " + strings.Join(queryLines, " ")
	if len(query) > 500 {
		query = query[:500]
	}

	return v.SearchCommits(query, limit)
}

// Helper functions

func extractFilesFromDiff(diff string) []string {
	var files []string
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[2], "a/")
				if file != "" && !contains(files, file) {
					files = append(files, file)
				}
			}
		}
	}
	return files
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
