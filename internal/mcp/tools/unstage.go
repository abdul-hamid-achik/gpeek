package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UnstageInput is the input for gpeek_unstage
type UnstageInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=File path to unstage (relative to repo root). If empty unstages all files."`
}

// RegisterUnstageTool registers the gpeek_unstage tool
func RegisterUnstageTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_unstage",
		Description: "Unstage files from the staging area. If path is empty, unstages all files.",
	}, handleUnstage)
}

func handleUnstage(ctx context.Context, req *mcp.CallToolRequest, input UnstageInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Path != "" {
		if err := ValidatePath(input.Path); err != nil {
			return ErrorResult(fmt.Sprintf("Invalid path: %v", err)), nil, nil
		}
		if err := repo.Unstage(input.Path); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to unstage file: %v", err)), nil, nil
		}
	} else {
		if err := repo.UnstageAll(); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to unstage all files: %v", err)), nil, nil
		}
	}

	// Get updated status to report what remains staged
	status, err := repo.Status()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Unstaged successfully but failed to get status: %v", err)), nil, nil
	}

	response := UnstageResponse{
		Remaining: convertFileEntries(status.Staged),
		Total:     len(status.Staged),
	}

	return ResultJSON(response)
}
