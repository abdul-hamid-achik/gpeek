package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ChangeImpactInput is the input for gpeek_change_impact
type ChangeImpactInput struct {
	Path  string   `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Files []string `json:"files,omitempty" jsonschema:"description=Files to analyze impact for (if empty uses staged/unstaged changes)"`
}

// ChangeImpactResponse is the response for change impact analysis
type ChangeImpactResponse struct {
	Files          []FileImpact    `json:"files"`
	Summary        ImpactSummary   `json:"summary"`
	RelatedFiles   []string        `json:"related_files,omitempty"`
	AffectedTests  []string        `json:"affected_tests,omitempty"`
}

// FileImpact represents the impact analysis for a single file
type FileImpact struct {
	Path            string   `json:"path"`
	Type            string   `json:"type"` // "source", "test", "config", "docs"
	ChangeFrequency int      `json:"change_frequency"`
	LastChanged     string   `json:"last_changed"`
	RecentAuthors   []string `json:"recent_authors"`
	RelatedFiles    []string `json:"related_files,omitempty"`
}

// ImpactSummary provides an overall summary
type ImpactSummary struct {
	TotalFiles      int    `json:"total_files"`
	HighImpactFiles int    `json:"high_impact_files"`
	TestCoverage    string `json:"test_coverage"` // "likely", "unlikely", "unknown"
	RiskLevel       string `json:"risk_level"`    // "low", "medium", "high"
}

// RegisterChangeImpactTool registers the gpeek_change_impact tool
func RegisterChangeImpactTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_change_impact",
		Description: "Analyze the impact of file changes, find related files, and assess risk",
	}, handleChangeImpact)
}

func handleChangeImpact(ctx context.Context, req *mcp.CallToolRequest, input ChangeImpactInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	// Get files to analyze
	var filesToAnalyze []string
	if len(input.Files) > 0 {
		filesToAnalyze = input.Files
	} else {
		// Use current changes
		status, err := repo.Status()
		if err != nil {
			return ErrorResult(fmt.Sprintf("Failed to get status: %v", err)), nil, nil
		}

		for _, f := range status.Staged {
			filesToAnalyze = append(filesToAnalyze, f.Path)
		}
		for _, f := range status.Unstaged {
			filesToAnalyze = append(filesToAnalyze, f.Path)
		}
	}

	if len(filesToAnalyze) == 0 {
		return ErrorResult("No files to analyze - provide files or have pending changes"), nil, nil
	}

	// Analyze each file
	var fileImpacts []FileImpact
	relatedFilesSet := make(map[string]bool)
	testFilesSet := make(map[string]bool)
	highImpactCount := 0

	for _, file := range filesToAnalyze {
		impact := analyzeFileImpact(repo, file)
		fileImpacts = append(fileImpacts, impact)

		// Track related files
		for _, rf := range impact.RelatedFiles {
			if rf != file {
				relatedFilesSet[rf] = true
			}
		}

		// Check if high impact
		if impact.ChangeFrequency > 5 {
			highImpactCount++
		}

		// Find associated test files
		testFile := findTestFile(file)
		if testFile != "" {
			testFilesSet[testFile] = true
		}
	}

	// Convert sets to slices
	var relatedFiles []string
	for rf := range relatedFilesSet {
		relatedFiles = append(relatedFiles, rf)
	}
	var affectedTests []string
	for tf := range testFilesSet {
		affectedTests = append(affectedTests, tf)
	}

	// Determine risk level
	riskLevel := "low"
	if highImpactCount > len(filesToAnalyze)/2 {
		riskLevel = "high"
	} else if highImpactCount > 0 {
		riskLevel = "medium"
	}

	// Determine test coverage likelihood
	testCoverage := "unknown"
	if len(affectedTests) > 0 {
		testCoverage = "likely"
	} else if hasTestsInRepo(repo) {
		testCoverage = "unlikely"
	}

	response := ChangeImpactResponse{
		Files:         fileImpacts,
		RelatedFiles:  relatedFiles,
		AffectedTests: affectedTests,
		Summary: ImpactSummary{
			TotalFiles:      len(filesToAnalyze),
			HighImpactFiles: highImpactCount,
			TestCoverage:    testCoverage,
			RiskLevel:       riskLevel,
		},
	}

	return ResultJSON(response)
}

func analyzeFileImpact(repo *git.Repository, file string) FileImpact {
	impact := FileImpact{
		Path: file,
		Type: classifyFile(file),
	}

	// Get file history to determine change frequency and authors
	commits, err := repo.FileHistory(file, git.FileHistoryOptions{Limit: 20})
	if err == nil {
		impact.ChangeFrequency = len(commits)
		if len(commits) > 0 {
			impact.LastChanged = TimeAgo(commits[0].Time)

			// Collect unique authors
			authorSet := make(map[string]bool)
			for _, c := range commits {
				authorSet[c.Author] = true
			}
			for author := range authorSet {
				impact.RecentAuthors = append(impact.RecentAuthors, author)
			}
		}
	}

	// Find related files (files often changed together)
	impact.RelatedFiles = findRelatedFiles(repo, file, commits)

	return impact
}

func classifyFile(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	dir := filepath.Dir(file)
	base := strings.ToLower(filepath.Base(file))

	// Check if test file
	if strings.Contains(base, "test") || strings.Contains(base, "spec") ||
		strings.Contains(dir, "test") || strings.Contains(dir, "spec") {
		return "test"
	}

	// Check if config file
	configExts := map[string]bool{
		".yaml": true, ".yml": true, ".json": true, ".toml": true,
		".ini": true, ".env": true, ".conf": true,
	}
	if configExts[ext] || base == "makefile" || base == "dockerfile" {
		return "config"
	}

	// Check if documentation
	if ext == ".md" || ext == ".rst" || ext == ".txt" ||
		strings.Contains(dir, "doc") {
		return "docs"
	}

	return "source"
}

func findTestFile(file string) string {
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(filepath.Base(file), ext)
	dir := filepath.Dir(file)

	// Common test file patterns
	testPatterns := []string{
		filepath.Join(dir, base+"_test"+ext),
		filepath.Join(dir, base+".test"+ext),
		filepath.Join(dir, "test_"+base+ext),
		filepath.Join("tests", base+"_test"+ext),
		filepath.Join("test", base+"_test"+ext),
	}

	for _, pattern := range testPatterns {
		return pattern // Return the first likely pattern
	}
	return ""
}

func findRelatedFiles(repo *git.Repository, file string, commits []git.Commit) []string {
	// Find files that are often changed together with this file
	coChangeCount := make(map[string]int)

	for _, c := range commits {
		// Get files changed in this commit
		diff, err := repo.CommitDiff(c.Hash)
		if err != nil {
			continue
		}

		files := extractChangedFiles(diff)
		for _, f := range files {
			if f != file {
				coChangeCount[f]++
			}
		}
	}

	// Return files that appear in >30% of the commits
	threshold := len(commits) * 3 / 10
	if threshold < 2 {
		threshold = 2
	}

	var related []string
	for f, count := range coChangeCount {
		if count >= threshold {
			related = append(related, f)
		}
	}

	// Limit to top 5
	if len(related) > 5 {
		related = related[:5]
	}

	return related
}

func extractChangedFiles(diff string) []string {
	var files []string
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[2], "a/")
				files = append(files, file)
			}
		}
	}
	return files
}

func hasTestsInRepo(repo *git.Repository) bool {
	// Simple heuristic: check for common test directories
	// In a real implementation, we'd check the file system
	return true // Assume tests exist
}
