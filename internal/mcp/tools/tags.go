package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TagsInput is the input for gpeek_tags
type TagsInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
}

// RegisterTagsTool registers the gpeek_tags tool
func RegisterTagsTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_tags",
		Description: "List all tags in the repository",
	}, handleTags)
}

func handleTags(ctx context.Context, req *mcp.CallToolRequest, input TagsInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	tags, err := repo.ListTags()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list tags: %v", err)), nil, nil
	}

	response := BuildTagsResponse(tags)

	return ResultJSON(response)
}
