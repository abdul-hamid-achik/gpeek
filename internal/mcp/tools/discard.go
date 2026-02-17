package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DiscardInput is the input for gpeek_discard
type DiscardInput struct {
	Path    string `json:"path" jsonschema:"description=File path to discard changes for (relative to repo root; required),required"`
	Confirm bool   `json:"confirm" jsonschema:"description=Must be true to confirm this destructive operation (required),required"`
}

// RegisterDiscardTool registers the gpeek_discard tool
func RegisterDiscardTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_discard",
		Description: "Discard unstaged changes for a file. This is destructive and requires confirm: true.",
	}, handleDiscard)
}

func handleDiscard(ctx context.Context, req *mcp.CallToolRequest, input DiscardInput) (*mcp.CallToolResult, any, error) {
	if !input.Confirm {
		return ErrorResult("Destructive operation: you must set confirm to true to discard changes"), nil, nil
	}

	if input.Path == "" {
		return ErrorResult("Path is required for discard operation"), nil, nil
	}

	if err := ValidatePath(input.Path); err != nil {
		return ErrorResult(fmt.Sprintf("Invalid path: %v", err)), nil, nil
	}

	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if err := repo.Discard(input.Path); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to discard changes: %v", err)), nil, nil
	}

	response := DiscardResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully discarded changes for %s", input.Path),
		Path:    input.Path,
	}

	return ResultJSON(response)
}
