package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DiffInput is the input for gpeek_diff
type DiffInput struct {
	Path         string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	File         string `json:"file,omitempty" jsonschema:"description=Specific file to diff (optional)"`
	Staged       bool   `json:"staged,omitempty" jsonschema:"description=Show staged changes instead of working tree changes"`
	Commit       string `json:"commit,omitempty" jsonschema:"description=Show diff for a specific commit hash"`
	IncludeBlame bool   `json:"include_blame,omitempty" jsonschema:"description=Include blame info for changed lines (slower)"`
	ContextLines int    `json:"context_lines,omitempty" jsonschema:"description=Number of context lines around changes (default: 3)"`
}

// RegisterDiffTool registers the gpeek_diff tool
func RegisterDiffTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_diff",
		Description: "Get structured diff with parsed hunks. Can show staged changes, working tree changes, or specific commit diffs",
	}, handleDiff)
}

func handleDiff(ctx context.Context, req *mcp.CallToolRequest, input DiffInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	var rawDiff string

	if input.Commit != "" {
		rawDiff, err = repo.CommitDiff(input.Commit)
	} else if input.File != "" {
		rawDiff, err = repo.FileDiff(input.File, input.Staged)
	} else {
		// Get all diffs
		status, statusErr := repo.Status()
		if statusErr != nil {
			return ErrorResult(fmt.Sprintf("Failed to get status: %v", statusErr)), nil, nil
		}

		var files []string
		if input.Staged {
			for _, f := range status.Staged {
				files = append(files, f.Path)
			}
		} else {
			for _, f := range status.Unstaged {
				files = append(files, f.Path)
			}
		}

		var builder strings.Builder
		for _, f := range files {
			d, _ := repo.FileDiff(f, input.Staged)
			builder.WriteString(d)
		}
		rawDiff = builder.String()
	}

	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get diff: %v", err)), nil, nil
	}

	parsed := diff.Parse(rawDiff)
	response := BuildDiffResponse(parsed, input.File, input.Commit, input.Staged)

	// Add blame info if requested
	if input.IncludeBlame && input.File != "" {
		addBlameToResponse(repo, &response)
	}

	return ResultJSON(response)
}

// addBlameToResponse enriches diff response with blame info
func addBlameToResponse(repo *git.Repository, response *DiffResponse) {
	for i := range response.Files {
		fileDiff := &response.Files[i]
		if fileDiff.IsBinary || fileDiff.IsNew || fileDiff.IsDelete {
			continue
		}

		// Get blame for the file
		fileName := fileDiff.NewName
		if fileName == "" {
			fileName = fileDiff.OldName
		}

		lines, err := repo.BlameFile(fileName)
		if err != nil {
			continue
		}

		// Create a map of line number to blame info
		blameMap := make(map[int]*BlameInfo)
		for _, line := range lines {
			hash := line.Hash
			if len(hash) > 8 {
				hash = hash[:8]
			}
			blameMap[line.LineNum] = &BlameInfo{
				Author:  line.Author,
				Hash:    hash,
				TimeAgo: TimeAgo(line.Time),
			}
		}

		// Add blame info to removed lines (using old line numbers)
		for j := range fileDiff.Hunks {
			hunk := &fileDiff.Hunks[j]
			for k := range hunk.Lines {
				lineInfo := &hunk.Lines[k]
				if lineInfo.Type == "remove" && lineInfo.OldNumber > 0 {
					if blame, ok := blameMap[lineInfo.OldNumber]; ok {
						lineInfo.Blame = blame
					}
				}
			}
		}
	}
}
