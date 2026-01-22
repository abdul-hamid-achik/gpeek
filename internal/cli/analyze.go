package cli

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/agent"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/spf13/cobra"
)

// ConflictCheckResponse represents the JSON response for conflict check
type ConflictCheckResponse struct {
	Branch           string             `json:"branch"`
	Into             string             `json:"into"`
	WouldConflict    bool               `json:"would_conflict"`
	ConflictingFiles []ConflictFileInfo `json:"conflicting_files,omitempty"`
	SafeToMerge      bool               `json:"safe_to_merge"`
	Recommendation   string             `json:"recommendation"`
}

type ConflictFileInfo struct {
	Path       string `json:"path"`
	HasOverlap bool   `json:"has_overlap"`
}

// SummarizeChangesResponse represents the JSON response for summarize-changes
type SummarizeChangesResponse struct {
	From        string            `json:"from"`
	To          string            `json:"to"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	ChangeTypes map[string]int    `json:"change_types"`
	FilesByArea map[string]int    `json:"files_by_area"`
	Stats       ChangeStatsInfo   `json:"stats"`
	Commits     []CommitBriefInfo `json:"commits,omitempty"`
}

type ChangeStatsInfo struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	CommitCount  int `json:"commit_count"`
}

type CommitBriefInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author,omitempty"`
}

// AnalyzeChangesResponse represents the JSON response for analyze-changes
type AnalyzeChangesResponse struct {
	ChangeType     string          `json:"change_type"`
	Confidence     float64         `json:"confidence"`
	AffectedAreas  []string        `json:"affected_areas"`
	SimilarChanges []SimilarInfo   `json:"similar_changes,omitempty"`
	RiskLevel      string          `json:"risk_level"`
	RiskReasons    []string        `json:"risk_reasons,omitempty"`
	Suggestions    []string        `json:"suggestions,omitempty"`
}

type SimilarInfo struct {
	Hash       string  `json:"hash"`
	Message    string  `json:"message"`
	Similarity float64 `json:"similarity"`
}

var (
	conflictBranch string
	conflictInto   string
	summarizeFrom  string
	summarizeTo    string
	analyzeStaged  bool
)

var checkConflictsCmd = &cobra.Command{
	Use:   "check-conflicts",
	Short: "Check for potential merge conflicts",
	Long:  `Perform a dry-run check to see if merging a branch would cause conflicts.`,
	RunE:  runCheckConflicts,
}

var summarizeChangesCmd = &cobra.Command{
	Use:   "summarize-changes",
	Short: "Summarize changes between refs",
	Long:  `Generate a summary of changes between two git references (branches, tags, commits).`,
	RunE:  runSummarizeChanges,
}

var analyzeChangesCmd = &cobra.Command{
	Use:   "analyze-changes",
	Short: "Analyze staged or unstaged changes",
	Long:  `Analyze changes to determine change type, risk level, and provide suggestions.`,
	RunE:  runAnalyzeChanges,
}

func init() {
	checkConflictsCmd.Flags().StringVarP(&conflictBranch, "branch", "b", "", "Branch to check for conflicts (required)")
	checkConflictsCmd.Flags().StringVar(&conflictInto, "into", "", "Target branch to merge into (default: current branch)")
	_ = checkConflictsCmd.MarkFlagRequired("branch")

	summarizeChangesCmd.Flags().StringVar(&summarizeFrom, "from", "", "Starting reference (required)")
	summarizeChangesCmd.Flags().StringVar(&summarizeTo, "to", "HEAD", "Ending reference")
	_ = summarizeChangesCmd.MarkFlagRequired("from")

	analyzeChangesCmd.Flags().BoolVarP(&analyzeStaged, "staged", "s", true, "Analyze staged changes (default: true)")

	RootCmd.AddCommand(checkConflictsCmd)
	RootCmd.AddCommand(summarizeChangesCmd)
	RootCmd.AddCommand(analyzeChangesCmd)
}

func runCheckConflicts(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	analyzer := agent.NewAnalyzer(repo)
	result, err := analyzer.CheckConflicts(conflictBranch, conflictInto)
	if err != nil {
		return fmt.Errorf("conflict check failed: %w", err)
	}

	// Convert to response type
	conflictFiles := make([]ConflictFileInfo, len(result.ConflictingFiles))
	for i, f := range result.ConflictingFiles {
		conflictFiles[i] = ConflictFileInfo{
			Path:       f.Path,
			HasOverlap: f.HasOverlap,
		}
	}

	response := ConflictCheckResponse{
		Branch:           result.Branch,
		Into:             result.Into,
		WouldConflict:    result.WouldConflict,
		ConflictingFiles: conflictFiles,
		SafeToMerge:      result.SafeToMerge,
		Recommendation:   result.Recommendation,
	}

	output(response, formatConflictCheckPlain)
	return nil
}

func runSummarizeChanges(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	analyzer := agent.NewAnalyzer(repo)
	result, err := analyzer.SummarizeChanges(summarizeFrom, summarizeTo)
	if err != nil {
		return fmt.Errorf("summarize failed: %w", err)
	}

	// Convert to response type
	commits := make([]CommitBriefInfo, len(result.Commits))
	for i, c := range result.Commits {
		commits[i] = CommitBriefInfo{
			Hash:    c.Hash,
			Message: c.Message,
			Author:  c.Author,
		}
	}

	response := SummarizeChangesResponse{
		From:        result.From,
		To:          result.To,
		Title:       result.Title,
		Description: result.Description,
		ChangeTypes: result.ChangeTypes,
		FilesByArea: result.FilesByArea,
		Stats: ChangeStatsInfo{
			FilesChanged: result.Stats.FilesChanged,
			Additions:    result.Stats.Additions,
			Deletions:    result.Stats.Deletions,
			CommitCount:  result.Stats.CommitCount,
		},
		Commits: commits,
	}

	output(response, formatSummarizePlain)
	return nil
}

func runAnalyzeChanges(cmd *cobra.Command, args []string) error {
	repo, err := git.Open(GetPath())
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	analyzer := agent.NewAnalyzer(repo)
	result, err := analyzer.AnalyzeChanges(analyzeStaged)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Convert to response type
	similarChanges := make([]SimilarInfo, len(result.SimilarChanges))
	for i, s := range result.SimilarChanges {
		similarChanges[i] = SimilarInfo{
			Hash:       s.Hash,
			Message:    s.Message,
			Similarity: s.Similarity,
		}
	}

	response := AnalyzeChangesResponse{
		ChangeType:     string(result.ChangeType),
		Confidence:     result.Confidence,
		AffectedAreas:  result.AffectedAreas,
		SimilarChanges: similarChanges,
		RiskLevel:      string(result.RiskLevel),
		RiskReasons:    result.RiskReasons,
		Suggestions:    result.Suggestions,
	}

	output(response, formatAnalyzePlain)
	return nil
}

func formatConflictCheckPlain(data interface{}) string {
	response := data.(ConflictCheckResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Conflict check: %s -> %s\n\n", response.Branch, response.Into))

	if response.WouldConflict {
		sb.WriteString("WARNING: Merge would cause conflicts!\n\n")
		sb.WriteString("Conflicting files:\n")
		for _, f := range response.ConflictingFiles {
			sb.WriteString(fmt.Sprintf("  - %s\n", f.Path))
		}
	} else {
		sb.WriteString("No conflicts detected.\n")
	}

	sb.WriteString(fmt.Sprintf("\nRecommendation: %s\n", response.Recommendation))

	return sb.String()
}

func formatSummarizePlain(data interface{}) string {
	response := data.(SummarizeChangesResponse)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Changes from %s to %s\n", response.From, response.To))
	sb.WriteString(fmt.Sprintf("Title: %s\n", response.Title))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", response.Description))

	if len(response.ChangeTypes) > 0 {
		sb.WriteString("Change types:\n")
		for t, count := range response.ChangeTypes {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", t, count))
		}
		sb.WriteString("\n")
	}

	if len(response.FilesByArea) > 0 {
		sb.WriteString("Files by area:\n")
		for area, count := range response.FilesByArea {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", area, count))
		}
		sb.WriteString("\n")
	}

	if len(response.Commits) > 0 && len(response.Commits) <= 10 {
		sb.WriteString("Commits:\n")
		for _, c := range response.Commits {
			sb.WriteString(fmt.Sprintf("  %s %s\n", c.Hash, c.Message))
		}
	}

	return sb.String()
}

func formatAnalyzePlain(data interface{}) string {
	response := data.(AnalyzeChangesResponse)
	var sb strings.Builder

	sb.WriteString("Change Analysis\n\n")
	sb.WriteString(fmt.Sprintf("Type: %s (%.0f%% confidence)\n", response.ChangeType, response.Confidence*100))
	sb.WriteString(fmt.Sprintf("Risk Level: %s\n", response.RiskLevel))

	if len(response.AffectedAreas) > 0 {
		sb.WriteString(fmt.Sprintf("Affected Areas: %s\n", strings.Join(response.AffectedAreas, ", ")))
	}

	if len(response.RiskReasons) > 0 {
		sb.WriteString("\nRisk Factors:\n")
		for _, r := range response.RiskReasons {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	if len(response.Suggestions) > 0 {
		sb.WriteString("\nSuggestions:\n")
		for _, s := range response.Suggestions {
			sb.WriteString(fmt.Sprintf("  - %s\n", s))
		}
	}

	return sb.String()
}
