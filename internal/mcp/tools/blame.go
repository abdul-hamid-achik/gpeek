package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BlameInput is the input for gpeek_blame
type BlameInput struct {
	Path      string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	File      string `json:"file" jsonschema:"description=File to blame,required"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"description=Start line number (optional)"`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"description=End line number (optional)"`
}

// RegisterBlameTool registers the gpeek_blame tool
func RegisterBlameTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_blame",
		Description: "Get line-by-line blame information showing who last modified each line",
	}, handleBlame)
}

func handleBlame(ctx context.Context, req *mcp.CallToolRequest, input BlameInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.File == "" {
		return ErrorResult("file parameter is required"), nil, nil
	}

	lines, err := repo.BlameFile(input.File)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get blame: %v", err)), nil, nil
	}

	// Apply line range filter
	if input.StartLine > 0 || input.EndLine > 0 {
		start := input.StartLine
		if start < 1 {
			start = 1
		}
		end := input.EndLine
		if end < 1 || end > len(lines) {
			end = len(lines)
		}
		if start <= end && start <= len(lines) {
			lines = lines[start-1 : end]
		}
	}

	response := BuildBlameResponse(input.File, lines)

	return ResultJSON(response)
}
