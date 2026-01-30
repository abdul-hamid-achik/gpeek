package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CodeOwnersInput is the input for gpeek_code_owners
type CodeOwnersInput struct {
	Path  string   `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Files []string `json:"files,omitempty" jsonschema:"description=Specific files to get owners for (if empty returns all CODEOWNERS rules)"`
}

// CodeOwnersResponse is the response for code owners queries
type CodeOwnersResponse struct {
	HasCodeowners bool                 `json:"has_codeowners"`
	CodeownersPath string              `json:"codeowners_path,omitempty"`
	FileOwners    []FileOwnerInfo      `json:"file_owners,omitempty"`
	AllRules      []CodeOwnerRuleInfo  `json:"all_rules,omitempty"`
}

// FileOwnerInfo represents ownership info for a file
type FileOwnerInfo struct {
	File    string   `json:"file"`
	Owners  []string `json:"owners"`
	Source  string   `json:"source"` // "codeowners" or "blame"
}

// CodeOwnerRuleInfo represents a CODEOWNERS rule
type CodeOwnerRuleInfo struct {
	Pattern string   `json:"pattern"`
	Owners  []string `json:"owners"`
	Line    int      `json:"line"`
}

// RegisterCodeOwnersTool registers the gpeek_code_owners tool
func RegisterCodeOwnersTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_code_owners",
		Description: "Get code ownership information from CODEOWNERS file or inferred from blame",
	}, handleCodeOwners)
}

func handleCodeOwners(ctx context.Context, req *mcp.CallToolRequest, input CodeOwnersInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	response := CodeOwnersResponse{}

	// Get CODEOWNERS file
	cof, err := repo.GetCodeOwners()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to read CODEOWNERS: %v", err)), nil, nil
	}

	if cof != nil {
		response.HasCodeowners = true
		response.CodeownersPath = cof.Path

		// Convert rules
		response.AllRules = make([]CodeOwnerRuleInfo, len(cof.Rules))
		for i, r := range cof.Rules {
			response.AllRules[i] = CodeOwnerRuleInfo{
				Pattern: r.Pattern,
				Owners:  r.Owners,
				Line:    r.Line,
			}
		}
	}

	// If specific files requested, get their owners
	if len(input.Files) > 0 {
		response.FileOwners = make([]FileOwnerInfo, 0, len(input.Files))

		for _, file := range input.Files {
			// Verify file exists
			fullPath := filepath.Join(repo.Path(), file)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				continue
			}

			owners, source, err := repo.GetFileOwners(file)
			if err != nil {
				// Skip files that can't be analyzed
				continue
			}

			response.FileOwners = append(response.FileOwners, FileOwnerInfo{
				File:   file,
				Owners: owners,
				Source: source,
			})
		}
	}

	return ResultJSON(response)
}
