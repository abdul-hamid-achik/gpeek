package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StageInput is the input for gpeek_stage
type StageInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=File path to stage (relative to repo root). If empty stages all changes."`
}

// RegisterStageTool registers the gpeek_stage tool
func RegisterStageTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_stage",
		Description: "Stage files for commit. If path is empty, stages all changes.",
	}, handleStage)
}

func handleStage(ctx context.Context, req *mcp.CallToolRequest, input StageInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Path != "" {
		if err := ValidatePath(input.Path); err != nil {
			return ErrorResult(fmt.Sprintf("Invalid path: %v", err)), nil, nil
		}
		if err := repo.Stage(input.Path); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to stage file: %v", err)), nil, nil
		}
	} else {
		if err := repo.StageAll(); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to stage all files: %v", err)), nil, nil
		}
	}

	// Get updated status to report what was staged
	status, err := repo.Status()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Staged successfully but failed to get status: %v", err)), nil, nil
	}

	response := StageResponse{
		Staged: convertFileEntries(status.Staged),
		Total:  len(status.Staged),
	}

	return ResultJSON(response)
}
