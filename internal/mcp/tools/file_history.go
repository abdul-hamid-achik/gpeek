package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FileHistoryInput is the input for gpeek_file_history
type FileHistoryInput struct {
	Path   string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	File   string `json:"file" jsonschema:"description=File path to get history for,required"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=Maximum number of commits to return (default: 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"description=Number of commits to skip (for pagination)"`
}

// FileHistoryResponse is the response for file history
type FileHistoryResponse struct {
	File       string       `json:"file"`
	Commits    []CommitInfo `json:"commits"`
	Total      int          `json:"total"`
	Offset     int          `json:"offset"`
	Limit      int          `json:"limit"`
	HasMore    bool         `json:"has_more"`
	NextOffset int          `json:"next_offset,omitempty"`
}

// RegisterFileHistoryTool registers the gpeek_file_history tool
func RegisterFileHistoryTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_file_history",
		Description: "Get commit history for a specific file with pagination support",
	}, handleFileHistory)
}

func handleFileHistory(ctx context.Context, req *mcp.CallToolRequest, input FileHistoryInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.File == "" {
		return ErrorResult("file parameter is required"), nil, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	// Request one extra to detect if there are more results
	commits, err := repo.FileHistory(input.File, git.FileHistoryOptions{
		Limit:  limit + 1,
		Offset: input.Offset,
	})
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get file history: %v", err)), nil, nil
	}

	hasMore := len(commits) > limit
	if hasMore {
		commits = commits[:limit]
	}

	commitInfos := make([]CommitInfo, len(commits))
	for i, c := range commits {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		commitInfos[i] = CommitInfo{
			Hash:      c.Hash,
			ShortHash: shortHash,
			Message:   c.Message,
			Author:    c.Author,
			Email:     c.Email,
			Time:      c.Time,
			TimeAgo:   TimeAgo(c.Time),
			IsMerge:   c.IsMerge,
			Parents:   c.Parents,
		}
	}

	response := FileHistoryResponse{
		File:    input.File,
		Commits: commitInfos,
		Total:   len(commitInfos),
		Offset:  input.Offset,
		Limit:   limit,
		HasMore: hasMore,
	}

	if hasMore {
		response.NextOffset = input.Offset + limit
	}

	return ResultJSON(response)
}
