package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ChangesBetweenInput is the input for gpeek_changes_between
type ChangesBetweenInput struct {
	Path string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	From string `json:"from" jsonschema:"description=Starting reference (branch tag or commit),required"`
	To   string `json:"to,omitempty" jsonschema:"description=Ending reference (default: HEAD)"`
}

// ChangesBetweenResponse is the response for changes between refs
type ChangesBetweenResponse struct {
	From  string         `json:"from"`
	To    string         `json:"to"`
	Files []FileDiffInfo `json:"files"`
	Stats struct {
		FilesChanged int `json:"files_changed"`
	} `json:"stats"`
}

// RegisterChangesBetweenTool registers the gpeek_changes_between tool
func RegisterChangesBetweenTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_changes_between",
		Description: "Analyze changes between two git references (branches, tags, commits)",
	}, handleChangesBetween)
}

func handleChangesBetween(ctx context.Context, req *mcp.CallToolRequest, input ChangesBetweenInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.From == "" {
		return ErrorResult("from parameter is required"), nil, nil
	}

	to := input.To
	if to == "" {
		to = "HEAD"
	}

	rawDiff, err := repo.CommitDiff(fmt.Sprintf("%s..%s", input.From, to))
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get changes: %v", err)), nil, nil
	}

	parsed := diff.Parse(rawDiff)
	diffResponse := BuildDiffResponse(parsed, "", "", false)

	response := ChangesBetweenResponse{
		From:  input.From,
		To:    to,
		Files: diffResponse.Files,
	}
	response.Stats.FilesChanged = len(parsed.Files)

	return ResultJSON(response)
}
