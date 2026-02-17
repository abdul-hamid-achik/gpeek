package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StashDropInput is the input for gpeek_stash_drop
type StashDropInput struct {
	Index   int  `json:"index" jsonschema:"description=Stash index to drop (required),required"`
	Confirm bool `json:"confirm" jsonschema:"description=Must be true to confirm this destructive operation (required),required"`
}

// RegisterStashDropTool registers the gpeek_stash_drop tool
func RegisterStashDropTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_stash_drop",
		Description: "Drop (delete) a stash entry. This is destructive and requires confirm: true.",
	}, handleStashDrop)
}

func handleStashDrop(ctx context.Context, req *mcp.CallToolRequest, input StashDropInput) (*mcp.CallToolResult, any, error) {
	if !input.Confirm {
		return ErrorResult("Destructive operation: you must set confirm to true to drop a stash"), nil, nil
	}

	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Index < 0 {
		return ErrorResult("Stash index must be non-negative"), nil, nil
	}

	ref := fmt.Sprintf("stash@{%d}", input.Index)

	if err := repo.StashDrop(input.Index); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to drop stash: %v", err)), nil, nil
	}

	response := StashOpResponse{
		Success:   true,
		Message:   fmt.Sprintf("Successfully dropped %s", ref),
		Reference: ref,
	}

	return ResultJSON(response)
}
