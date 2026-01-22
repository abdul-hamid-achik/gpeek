package agent

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
)

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeBugFix     ChangeType = "bug_fix"
	ChangeTypeFeature    ChangeType = "feature"
	ChangeTypeRefactor   ChangeType = "refactor"
	ChangeTypeDocs       ChangeType = "docs"
	ChangeTypeTest       ChangeType = "test"
	ChangeTypeStyle      ChangeType = "style"
	ChangeTypeChore      ChangeType = "chore"
	ChangeTypePerf       ChangeType = "performance"
	ChangeTypeUnknown    ChangeType = "unknown"
)

// RiskLevel represents the risk level of changes
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// ConflictInfo represents potential merge conflict information
type ConflictInfo struct {
	Path         string   `json:"path"`
	OurChanges   []int    `json:"our_changes,omitempty"`
	TheirChanges []int    `json:"their_changes,omitempty"`
	HasOverlap   bool     `json:"has_overlap"`
}

// ConflictCheckResult represents the result of a conflict check
type ConflictCheckResult struct {
	Branch           string         `json:"branch"`
	Into             string         `json:"into"`
	WouldConflict    bool           `json:"would_conflict"`
	ConflictingFiles []ConflictInfo `json:"conflicting_files,omitempty"`
	SafeToMerge      bool           `json:"safe_to_merge"`
	Recommendation   string         `json:"recommendation"`
}

// ChangeSummary represents a summary of changes between refs
type ChangeSummary struct {
	From        string            `json:"from"`
	To          string            `json:"to"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	ChangeTypes map[string]int    `json:"change_types"`
	FilesByArea map[string]int    `json:"files_by_area"`
	Stats       ChangeStats       `json:"stats"`
	Commits     []CommitSummary   `json:"commits,omitempty"`
}

// ChangeStats contains statistics about changes
type ChangeStats struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	CommitCount  int `json:"commit_count"`
}

// CommitSummary is a brief commit summary
type CommitSummary struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
}

// AnalysisResult represents the result of change analysis
type AnalysisResult struct {
	ChangeType        ChangeType        `json:"change_type"`
	Confidence        float64           `json:"confidence"`
	AffectedAreas     []string          `json:"affected_areas"`
	SimilarChanges    []SimilarChange   `json:"similar_changes,omitempty"`
	RiskLevel         RiskLevel         `json:"risk_level"`
	RiskReasons       []string          `json:"risk_reasons,omitempty"`
	Suggestions       []string          `json:"suggestions,omitempty"`
}

// SimilarChange represents a similar historical change
type SimilarChange struct {
	Hash       string  `json:"hash"`
	Message    string  `json:"message"`
	Similarity float64 `json:"similarity"`
}

// Analyzer provides smart analysis capabilities
type Analyzer struct {
	repo *git.Repository
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(repo *git.Repository) *Analyzer {
	return &Analyzer{repo: repo}
}

// CheckConflicts performs a dry-run merge conflict check
func (a *Analyzer) CheckConflicts(branch, into string) (*ConflictCheckResult, error) {
	if into == "" {
		into = a.repo.CurrentBranch()
	}

	result := &ConflictCheckResult{
		Branch: branch,
		Into:   into,
	}

	// Try a dry-run merge using git
	cmd := exec.Command("git", "merge-tree", "--write-tree", into, branch)
	cmd.Dir = a.repo.Path()
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's a conflict error
		outputStr := string(output)
		if strings.Contains(outputStr, "CONFLICT") || strings.Contains(outputStr, "conflict") {
			result.WouldConflict = true
			result.SafeToMerge = false
			result.Recommendation = fmt.Sprintf("Resolve conflicts in branch '%s' before merging into '%s'", branch, into)

			// Parse conflict information
			conflicts := parseConflicts(outputStr)
			result.ConflictingFiles = conflicts
		} else {
			// Some other error
			return nil, fmt.Errorf("merge check failed: %s", outputStr)
		}
	} else {
		result.WouldConflict = false
		result.SafeToMerge = true
		result.Recommendation = fmt.Sprintf("Safe to merge '%s' into '%s'", branch, into)
	}

	return result, nil
}

// SummarizeChanges provides a summary of changes between two refs
func (a *Analyzer) SummarizeChanges(from, to string) (*ChangeSummary, error) {
	if to == "" {
		to = "HEAD"
	}

	summary := &ChangeSummary{
		From:        from,
		To:          to,
		ChangeTypes: make(map[string]int),
		FilesByArea: make(map[string]int),
	}

	// Get commits between refs
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("%s..%s", from, to))
	cmd.Dir = a.repo.Path()
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]

			summary.Commits = append(summary.Commits, CommitSummary{
				Hash:    hash,
				Message: message,
			})

			// Categorize by change type
			changeType := categorizeCommit(message)
			summary.ChangeTypes[string(changeType)]++
		}
	}

	summary.Stats.CommitCount = len(summary.Commits)

	// Get diff stats
	rawDiff, err := a.repo.CommitDiff(fmt.Sprintf("%s..%s", from, to))
	if err == nil {
		parsed := diff.Parse(rawDiff)
		summary.Stats.FilesChanged = len(parsed.Files)
		added, removed := parsed.Stats()
		summary.Stats.Additions = added
		summary.Stats.Deletions = removed

		// Categorize files by area
		for _, f := range parsed.Files {
			area := extractArea(f.NewName)
			summary.FilesByArea[area]++
		}
	}

	// Generate title and description
	summary.Title = generateTitle(summary)
	summary.Description = generateDescription(summary)

	return summary, nil
}

// AnalyzeChanges analyzes staged changes and provides insights
func (a *Analyzer) AnalyzeChanges(staged bool) (*AnalysisResult, error) {
	status, err := a.repo.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []git.FileEntry
	if staged {
		files = status.Staged
	} else {
		files = status.Unstaged
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no changes to analyze")
	}

	result := &AnalysisResult{
		ChangeType:    ChangeTypeUnknown,
		Confidence:    0.5,
		AffectedAreas: []string{},
		RiskLevel:     RiskLevelLow,
	}

	// Analyze file types and areas
	areas := make(map[string]bool)
	for _, f := range files {
		area := extractArea(f.Path)
		areas[area] = true
	}
	for area := range areas {
		result.AffectedAreas = append(result.AffectedAreas, area)
	}

	// Determine change type based on files and content
	result.ChangeType, result.Confidence = determineChangeType(files)

	// Assess risk level
	result.RiskLevel, result.RiskReasons = assessRisk(files, result.AffectedAreas)

	// Add suggestions
	result.Suggestions = generateSuggestions(result)

	return result, nil
}

// Helper functions

func parseConflicts(output string) []ConflictInfo {
	var conflicts []ConflictInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "CONFLICT") {
			// Extract file path from conflict message
			// Format: CONFLICT (content): Merge conflict in <path>
			if idx := strings.Index(line, "Merge conflict in "); idx != -1 {
				path := strings.TrimSpace(line[idx+len("Merge conflict in "):])
				conflicts = append(conflicts, ConflictInfo{
					Path:       path,
					HasOverlap: true,
				})
			}
		}
	}

	return conflicts
}

func categorizeCommit(message string) ChangeType {
	msgLower := strings.ToLower(message)

	if strings.HasPrefix(msgLower, "fix") || strings.Contains(msgLower, "bug") || strings.Contains(msgLower, "issue") {
		return ChangeTypeBugFix
	}
	if strings.HasPrefix(msgLower, "feat") || strings.Contains(msgLower, "add") || strings.Contains(msgLower, "implement") {
		return ChangeTypeFeature
	}
	if strings.HasPrefix(msgLower, "refactor") || strings.Contains(msgLower, "cleanup") || strings.Contains(msgLower, "reorganize") {
		return ChangeTypeRefactor
	}
	if strings.HasPrefix(msgLower, "doc") || strings.Contains(msgLower, "readme") || strings.Contains(msgLower, "comment") {
		return ChangeTypeDocs
	}
	if strings.HasPrefix(msgLower, "test") || strings.Contains(msgLower, "_test") || strings.Contains(msgLower, "spec") {
		return ChangeTypeTest
	}
	if strings.HasPrefix(msgLower, "style") || strings.Contains(msgLower, "format") || strings.Contains(msgLower, "lint") {
		return ChangeTypeStyle
	}
	if strings.HasPrefix(msgLower, "chore") || strings.Contains(msgLower, "update dep") || strings.Contains(msgLower, "bump") {
		return ChangeTypeChore
	}
	if strings.HasPrefix(msgLower, "perf") || strings.Contains(msgLower, "optim") || strings.Contains(msgLower, "speed") {
		return ChangeTypePerf
	}

	return ChangeTypeUnknown
}

func extractArea(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return parts[0] + "/"
	}
	// Return file extension as area
	if idx := strings.LastIndex(path, "."); idx != -1 {
		return "*" + path[idx:]
	}
	return "root"
}

func generateTitle(summary *ChangeSummary) string {
	if len(summary.Commits) == 0 {
		return "No changes"
	}
	if len(summary.Commits) == 1 {
		return summary.Commits[0].Message
	}

	// Find the dominant change type
	maxType := ""
	maxCount := 0
	for t, count := range summary.ChangeTypes {
		if count > maxCount {
			maxType = t
			maxCount = count
		}
	}

	if maxType != "" && maxType != string(ChangeTypeUnknown) {
		return fmt.Sprintf("%s changes (%d commits)", maxType, len(summary.Commits))
	}

	return fmt.Sprintf("Multiple changes (%d commits)", len(summary.Commits))
}

func generateDescription(summary *ChangeSummary) string {
	var parts []string

	if summary.Stats.CommitCount > 0 {
		parts = append(parts, fmt.Sprintf("%d commits", summary.Stats.CommitCount))
	}
	if summary.Stats.FilesChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d files changed", summary.Stats.FilesChanged))
	}
	if summary.Stats.Additions > 0 {
		parts = append(parts, fmt.Sprintf("+%d additions", summary.Stats.Additions))
	}
	if summary.Stats.Deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d deletions", summary.Stats.Deletions))
	}

	return strings.Join(parts, ", ")
}

func determineChangeType(files []git.FileEntry) (ChangeType, float64) {
	hasTest := false
	hasDoc := false
	hasCode := false

	for _, f := range files {
		path := strings.ToLower(f.Path)
		if strings.Contains(path, "_test.") || strings.Contains(path, "/test/") || strings.Contains(path, "spec/") {
			hasTest = true
		}
		if strings.HasSuffix(path, ".md") || strings.Contains(path, "doc") || strings.Contains(path, "readme") {
			hasDoc = true
		}
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".py") || strings.HasSuffix(path, ".rs") {
			hasCode = true
		}
	}

	if hasTest && !hasCode {
		return ChangeTypeTest, 0.8
	}
	if hasDoc && !hasCode {
		return ChangeTypeDocs, 0.8
	}
	if hasCode {
		return ChangeTypeFeature, 0.6
	}

	return ChangeTypeUnknown, 0.5
}

func assessRisk(files []git.FileEntry, areas []string) (RiskLevel, []string) {
	var reasons []string
	risk := RiskLevelLow

	// High risk patterns
	for _, f := range files {
		path := strings.ToLower(f.Path)

		if strings.Contains(path, "auth") || strings.Contains(path, "security") {
			reasons = append(reasons, "touches authentication/security code")
			risk = RiskLevelHigh
		}
		if strings.Contains(path, "database") || strings.Contains(path, "migration") {
			reasons = append(reasons, "includes database changes")
			if risk < RiskLevelMedium {
				risk = RiskLevelMedium
			}
		}
		if strings.Contains(path, "config") || strings.Contains(path, ".env") {
			reasons = append(reasons, "modifies configuration")
			if risk < RiskLevelMedium {
				risk = RiskLevelMedium
			}
		}
	}

	// Large changes are higher risk
	if len(files) > 10 {
		reasons = append(reasons, fmt.Sprintf("large change (%d files)", len(files)))
		if risk < RiskLevelMedium {
			risk = RiskLevelMedium
		}
	}

	// Multiple areas are higher risk
	if len(areas) > 3 {
		reasons = append(reasons, fmt.Sprintf("affects multiple areas (%d)", len(areas)))
		if risk < RiskLevelMedium {
			risk = RiskLevelMedium
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "standard change pattern")
	}

	return risk, reasons
}

func generateSuggestions(result *AnalysisResult) []string {
	var suggestions []string

	if result.RiskLevel == RiskLevelHigh {
		suggestions = append(suggestions, "Consider requesting additional code review")
		suggestions = append(suggestions, "Ensure comprehensive test coverage")
	}

	if result.ChangeType == ChangeTypeFeature {
		suggestions = append(suggestions, "Add or update tests for new functionality")
		suggestions = append(suggestions, "Update documentation if needed")
	}

	if len(result.AffectedAreas) > 3 {
		suggestions = append(suggestions, "Consider splitting into smaller changes")
	}

	return suggestions
}
