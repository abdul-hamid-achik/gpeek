package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LogInput is the input for gpeek_log
type LogInput struct {
	Path   string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=Maximum number of commits to return (default: 50)"`
	Author string `json:"author,omitempty" jsonschema:"description=Filter commits by author name"`
	Since  string `json:"since,omitempty" jsonschema:"description=Show commits since date (e.g. '2024-01-01')"`
}

// RegisterLogTool registers the gpeek_log tool
func RegisterLogTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_log",
		Description: "Get commit history with optional filters",
	}, handleLog)
}

func handleLog(ctx context.Context, req *mcp.CallToolRequest, input LogInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	commits, err := repo.Log(limit)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get log: %v", err)), nil, nil
	}

	response := BuildLogResponse(commits)

	return ResultJSON(response)
}
