package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// CodeOwnerRule represents a single CODEOWNERS rule
type CodeOwnerRule struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
	Line    int      `json:"line"`
}

// CodeOwnersFile represents a parsed CODEOWNERS file
type CodeOwnersFile struct {
	Path  string          `json:"path"`
	Rules []CodeOwnerRule `json:"rules"`
}

// GetCodeOwners returns the CODEOWNERS file if it exists
func (r *Repository) GetCodeOwners() (*CodeOwnersFile, error) {
	// Check standard CODEOWNERS locations
	locations := []string{
		"CODEOWNERS",
		".github/CODEOWNERS",
		"docs/CODEOWNERS",
	}

	for _, loc := range locations {
		fullPath := filepath.Join(r.path, loc)
		if _, err := os.Stat(fullPath); err == nil {
			return parseCodeOwnersFile(fullPath, loc)
		}
	}

	return nil, nil // No CODEOWNERS file found
}

// parseCodeOwnersFile parses a CODEOWNERS file
func parseCodeOwnersFile(fullPath, relativePath string) (*CodeOwnersFile, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cof := &CodeOwnersFile{
		Path:  relativePath,
		Rules: []CodeOwnerRule{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line: pattern owner1 owner2 ...
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // Need at least pattern and one owner
		}

		rule := CodeOwnerRule{
			Pattern: parts[0],
			Owners:  parts[1:],
			Line:    lineNum,
		}
		cof.Rules = append(cof.Rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cof, nil
}

// FindOwners finds the owners for a given file path
func (cof *CodeOwnersFile) FindOwners(path string) []string {
	if cof == nil || len(cof.Rules) == 0 {
		return nil
	}

	// CODEOWNERS uses the last matching rule
	var matchedOwners []string

	for _, rule := range cof.Rules {
		if matchCodeOwnersPattern(rule.Pattern, path) {
			matchedOwners = rule.Owners
		}
	}

	return matchedOwners
}

// matchCodeOwnersPattern matches a CODEOWNERS pattern against a file path
func matchCodeOwnersPattern(pattern, path string) bool {
	// Normalize paths
	path = strings.TrimPrefix(path, "/")
	pattern = strings.TrimPrefix(pattern, "/")

	// Handle exact matches
	if pattern == path {
		return true
	}

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		dir := strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(path, dir+"/") || path == dir
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Convert to a simple glob-like match
		return globMatch(pattern, path)
	}

	// Handle directory match (pattern without /)
	if !strings.Contains(pattern, "/") {
		// Match any file with this name in any directory
		baseName := filepath.Base(path)
		if pattern == baseName {
			return true
		}
		// Also match as directory prefix
		return strings.HasPrefix(path, pattern+"/")
	}

	// Handle path prefix match
	return strings.HasPrefix(path, pattern+"/") || path == pattern
}

// globMatch performs simple glob matching
func globMatch(pattern, path string) bool {
	// Handle ** (match any path)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}
			if suffix != "" && !strings.HasSuffix(path, suffix) {
				return false
			}
			return true
		}
	}

	// Handle single * (match within a segment)
	if strings.Contains(pattern, "*") {
		// Simple prefix/suffix matching
		idx := strings.Index(pattern, "*")
		prefix := pattern[:idx]
		suffix := pattern[idx+1:]

		if prefix != "" && !strings.HasPrefix(path, prefix) {
			return false
		}
		if suffix != "" && !strings.HasSuffix(path, suffix) {
			return false
		}
		return true
	}

	return false
}

// GetFileOwners returns the owners for a specific file
// Falls back to blame-based heuristics if no CODEOWNERS rule matches
func (r *Repository) GetFileOwners(path string) ([]string, string, error) {
	// First try CODEOWNERS
	cof, err := r.GetCodeOwners()
	if err != nil {
		return nil, "", err
	}

	if cof != nil {
		owners := cof.FindOwners(path)
		if len(owners) > 0 {
			return owners, "codeowners", nil
		}
	}

	// Fallback to blame-based heuristics
	return r.getOwnersFromBlame(path)
}

// getOwnersFromBlame returns top contributors based on blame
func (r *Repository) getOwnersFromBlame(path string) ([]string, string, error) {
	lines, err := r.BlameFile(path)
	if err != nil {
		return nil, "", err
	}

	// Count contributions by author
	authorCounts := make(map[string]int)
	for _, line := range lines {
		if line.Author != "" {
			authorCounts[line.Author]++
		}
	}

	// Find top contributors (those with >10% of lines)
	totalLines := len(lines)
	if totalLines == 0 {
		return nil, "blame", nil
	}

	threshold := totalLines / 10
	if threshold < 1 {
		threshold = 1
	}

	var owners []string
	for author, count := range authorCounts {
		if count >= threshold {
			owners = append(owners, author)
		}
	}

	// Sort by contribution count (most contributions first)
	// Simple bubble sort for small lists
	for i := 0; i < len(owners)-1; i++ {
		for j := i + 1; j < len(owners); j++ {
			if authorCounts[owners[j]] > authorCounts[owners[i]] {
				owners[i], owners[j] = owners[j], owners[i]
			}
		}
	}

	// Limit to top 5
	if len(owners) > 5 {
		owners = owners[:5]
	}

	return owners, "blame", nil
}
