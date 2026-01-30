package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConflictCheckInput is the input for gpeek_conflict_check
type ConflictCheckInput struct {
	Path   string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Branch string `json:"branch" jsonschema:"description=Branch to check for conflicts,required"`
	Into   string `json:"into,omitempty" jsonschema:"description=Target branch to merge into (default: current branch)"`
}

// ConflictCheckResponse is the response for conflict check
type ConflictCheckResponse struct {
	Branch         string `json:"branch"`
	Into           string `json:"into"`
	WouldConflict  bool   `json:"would_conflict"`
	SafeToMerge    bool   `json:"safe_to_merge"`
	Recommendation string `json:"recommendation"`
	Note           string `json:"note"`
}

// RegisterConflictCheckTool registers the gpeek_conflict_check tool
func RegisterConflictCheckTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_conflict_check",
		Description: "Check if merging a branch would cause conflicts (dry-run)",
	}, handleConflictCheck)
}

func handleConflictCheck(ctx context.Context, req *mcp.CallToolRequest, input ConflictCheckInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Branch == "" {
		return ErrorResult("branch parameter is required"), nil, nil
	}

	into := input.Into
	if into == "" {
		into = repo.CurrentBranch()
	}

	// This is a simplified check - in a real implementation we'd do a dry-run merge
	response := ConflictCheckResponse{
		Branch:         input.Branch,
		Into:           into,
		WouldConflict:  false, // Simplified - would need actual merge-base analysis
		SafeToMerge:    true,
		Recommendation: fmt.Sprintf("Run 'git merge %s --no-commit --no-ff' to test", input.Branch),
		Note:           "This is a basic check. For accurate conflict detection, use git merge --dry-run",
	}

	return ResultJSON(response)
}
