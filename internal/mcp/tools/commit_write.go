package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CommitWriteInput is the input for gpeek_commit
type CommitWriteInput struct {
	Message string `json:"message" jsonschema:"description=Commit message (required),required"`
}

// RegisterCommitWriteTool registers the gpeek_commit tool
func RegisterCommitWriteTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_commit",
		Description: "Create a git commit with the currently staged changes.",
	}, handleCommitWrite)
}

func handleCommitWrite(ctx context.Context, req *mcp.CallToolRequest, input CommitWriteInput) (*mcp.CallToolResult, any, error) {
	if input.Message == "" {
		return ErrorResult("Commit message is required"), nil, nil
	}

	repo, err := OpenRepo("")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	// Check that there are staged changes
	status, err := repo.Status()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get status: %v", err)), nil, nil
	}
	if len(status.Staged) == 0 {
		return ErrorResult("Nothing to commit: no staged changes"), nil, nil
	}

	if err := repo.Commit(input.Message); err != nil {
		return ErrorResult(fmt.Sprintf("Failed to create commit: %v", err)), nil, nil
	}

	// Get the commit info
	message, hash, err := repo.GetLastCommitInfo()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Committed but failed to get commit info: %v", err)), nil, nil
	}

	shortHash := hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	response := CommitWriteResponse{
		Hash:      hash,
		ShortHash: shortHash,
		Message:   message,
	}

	return ResultJSON(response)
}
