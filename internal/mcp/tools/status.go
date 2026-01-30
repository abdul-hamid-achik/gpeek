package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StatusInput is the input for gpeek_status
type StatusInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
}

// StatusResponse is the response for status queries
type StatusResponse struct {
	Repository RepositoryInfo `json:"repository"`
	Staged     []FileInfo     `json:"staged"`
	Unstaged   []FileInfo     `json:"unstaged"`
	Untracked  []string       `json:"untracked"`
	Summary    StatusSummary  `json:"summary"`
}

// StatusSummary provides summary counts
type StatusSummary struct {
	StagedCount    int  `json:"staged_count"`
	UnstagedCount  int  `json:"unstaged_count"`
	UntrackedCount int  `json:"untracked_count"`
	IsClean        bool `json:"is_clean"`
	HasConflicts   bool `json:"has_conflicts"`
}

// RegisterStatusTool registers the gpeek_status tool
func RegisterStatusTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_status",
		Description: "Get comprehensive git repository status including staged, unstaged, and untracked files",
	}, handleStatus)
}

func handleStatus(ctx context.Context, req *mcp.CallToolRequest, input StatusInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	status, err := repo.Status()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get status: %v", err)), nil, nil
	}

	response := StatusResponse{
		Repository: RepositoryInfo{
			Name:   repo.Name(),
			Path:   repo.Path(),
			Branch: repo.CurrentBranch(),
		},
		Staged:    convertFileEntries(status.Staged),
		Unstaged:  convertFileEntries(status.Unstaged),
		Untracked: status.Untracked,
		Summary: StatusSummary{
			StagedCount:    len(status.Staged),
			UnstagedCount:  len(status.Unstaged),
			UntrackedCount: len(status.Untracked),
			IsClean:        len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
			HasConflicts:   hasConflicts(status),
		},
	}

	return ResultJSON(response)
}

func convertFileEntries(entries []git.FileEntry) []FileInfo {
	result := make([]FileInfo, len(entries))
	for i, e := range entries {
		result[i] = FileInfo{
			Path:   e.Path,
			Status: e.Status.String(),
		}
	}
	return result
}

func hasConflicts(status *git.Status) bool {
	for _, f := range status.Staged {
		if f.Status == git.StatusUpdatedButUnmerged {
			return true
		}
	}
	for _, f := range status.Unstaged {
		if f.Status == git.StatusUpdatedButUnmerged {
			return true
		}
	}
	return false
}
