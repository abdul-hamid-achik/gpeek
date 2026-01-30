package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BranchesInput is the input for gpeek_branches
type BranchesInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	All  bool   `json:"all,omitempty" jsonschema:"description=Include remote branches"`
}

// RegisterBranchesTool registers the gpeek_branches tool
func RegisterBranchesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_branches",
		Description: "List local and optionally remote branches",
	}, handleBranches)
}

func handleBranches(ctx context.Context, req *mcp.CallToolRequest, input BranchesInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	localBranches, err := repo.ListBranches()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list branches: %v", err)), nil, nil
	}

	var allBranches []git.Branch
	allBranches = append(allBranches, localBranches...)

	if input.All {
		remoteBranches, _ := repo.ListRemoteBranches()
		allBranches = append(allBranches, remoteBranches...)
	}

	response := BuildBranchesResponse(repo.CurrentBranch(), allBranches)

	return ResultJSON(response)
}
