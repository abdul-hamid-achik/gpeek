package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StashPopInput is the input for gpeek_stash_pop
type StashPopInput struct {
	Index int `json:"index,omitempty" jsonschema:"description=Stash index to pop (default: 0)"`
}

// RegisterStashPopTool registers the gpeek_stash_pop tool
func RegisterStashPopTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_stash_pop",
		Description: "Pop a stash entry, applying it to the working directory and removing it from the stash list.",
	}, handleStashPop)
}

func handleStashPop(ctx context.Context, req *mcp.CallToolRequest, input StashPopInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Index < 0 {
		return ErrorResult("Stash index must be non-negative"), nil, nil
	}

	ref := fmt.Sprintf("stash@{%d}", input.Index)

	if err := repo.StashPop(input.Index); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to pop stash: %v", err)), nil, nil
	}

	response := StashOpResponse{
		Success:   true,
		Message:   fmt.Sprintf("Successfully popped %s", ref),
		Reference: ref,
	}

	return ResultJSON(response)
}
