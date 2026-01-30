package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StashesInput is the input for gpeek_stashes
type StashesInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
}

// RegisterStashesTool registers the gpeek_stashes tool
func RegisterStashesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_stashes",
		Description: "List all stashed changes",
	}, handleStashes)
}

func handleStashes(ctx context.Context, req *mcp.CallToolRequest, input StashesInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	stashes, err := repo.StashList()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list stashes: %v", err)), nil, nil
	}

	response := BuildStashesResponse(stashes)

	return ResultJSON(response)
}
