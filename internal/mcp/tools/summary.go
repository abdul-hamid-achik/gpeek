package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SummaryInput is the input for gpeek_summary
type SummaryInput struct {
	Path     string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Commits  int    `json:"commits,omitempty" jsonschema:"description=Number of recent commits to include (default: 10)"`
	Enhanced bool   `json:"enhanced,omitempty" jsonschema:"description=Include hot files, languages, and project type detection"`
}

// RegisterSummaryTool registers the gpeek_summary tool
func RegisterSummaryTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_summary",
		Description: "Get complete repository snapshot in one call - ideal for understanding full repo context. Includes status, recent commits, branches, stashes, and tags",
	}, handleSummary)
}

func handleSummary(ctx context.Context, req *mcp.CallToolRequest, input SummaryInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	commitLimit := input.Commits
	if commitLimit <= 0 {
		commitLimit = 10
	}

	response := BuildSummaryResponse(repo, commitLimit)

	// Add enhanced analysis if requested
	if input.Enhanced {
		analysis, err := repo.AnalyzeRepository(50) // Analyze more commits for hot files
		if err == nil && analysis != nil {
			enhanced := &EnhancedSummary{
				ProjectType: analysis.ProjectType,
			}

			// Convert hot files
			enhanced.HotFiles = make([]HotFile, len(analysis.HotFiles))
			for i, hf := range analysis.HotFiles {
				enhanced.HotFiles[i] = HotFile{
					Path:        hf.Path,
					ChangeCount: hf.ChangeCount,
					Authors:     hf.Authors,
				}
			}

			// Convert languages
			enhanced.Languages = make([]LanguageInfo, len(analysis.Languages))
			for i, lang := range analysis.Languages {
				enhanced.Languages[i] = LanguageInfo{
					Name:       lang.Name,
					FileCount:  lang.FileCount,
					Percentage: lang.Percentage,
				}
			}

			// Add suggestions based on analysis
			enhanced.Suggestions = generateSuggestions(repo, &response, analysis)

			response.Enhanced = enhanced
		}
	}

	return ResultJSON(response)
}

// generateSuggestions creates actionable suggestions based on analysis
func generateSuggestions(_ interface{ CurrentBranch() string }, summary *SummaryResponse, analysis *git.RepoAnalysis) []string {
	var suggestions []string

	// Suggest based on status
	if summary.Status.StagedCount > 0 {
		suggestions = append(suggestions, "You have staged changes ready to commit")
	}
	if summary.Status.UnstagedCount > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Review %d unstaged changes", summary.Status.UnstagedCount))
	}
	if summary.Status.HasConflicts {
		suggestions = append(suggestions, "Resolve merge conflicts before proceeding")
	}

	// Suggest based on hot files
	if len(analysis.HotFiles) > 0 && analysis.HotFiles[0].ChangeCount > 10 {
		suggestions = append(suggestions,
			fmt.Sprintf("Consider reviewing %s - it has been changed %d times recently",
				analysis.HotFiles[0].Path, analysis.HotFiles[0].ChangeCount))
	}

	return suggestions
}
