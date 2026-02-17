package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StashSaveInput is the input for gpeek_stash_save
type StashSaveInput struct {
	Message string `json:"message,omitempty" jsonschema:"description=Optional stash message"`
}

// RegisterStashSaveTool registers the gpeek_stash_save tool
func RegisterStashSaveTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_stash_save",
		Description: "Save current changes to the stash.",
	}, handleStashSave)
}

func handleStashSave(ctx context.Context, req *mcp.CallToolRequest, input StashSaveInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if err := repo.StashSave(input.Message); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to save stash: %v", err)), nil, nil
	}

	// Get the latest stash to return its reference
	stashes, err := repo.StashList()
	if err != nil || len(stashes) == 0 {
		return ResultJSON(StashSaveResponse{
			Reference: "stash@{0}",
			Message:   input.Message,
		})
	}

	response := StashSaveResponse{
		Reference: fmt.Sprintf("stash@{%d}", stashes[0].Index),
		Message:   stashes[0].Message,
	}

	return ResultJSON(response)
}
