package tools

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DefaultPath returns "." if path is empty
func DefaultPath(path string) string {
	if path == "" {
		return "."
	}
	return path
}

// ValidatePath performs basic path validation/sandboxing
func ValidatePath(path string) error {
	cleaned := filepath.Clean(path)

	if strings.Contains(cleaned, "..") {
		abs, err := filepath.Abs(cleaned)
		if err != nil {
			return fmt.Errorf("invalid path: %v", err)
		}
		cwd, err := filepath.Abs(".")
		if err != nil {
			return fmt.Errorf("cannot determine current directory: %v", err)
		}
		if !strings.HasPrefix(abs, cwd) && cleaned != "." && !filepath.IsAbs(cleaned) {
			return fmt.Errorf("path traversal not allowed: %s", path)
		}
	}

	return nil
}

// ErrorResult creates an error result for tool handlers
func ErrorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}

// ResultJSON creates a successful JSON result
func ResultJSON(data any) (*mcp.CallToolResult, any, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}

// OpenRepo opens a repository with validation and caching
func OpenRepo(path string) (*git.Repository, error) {
	path = DefaultPath(path)
	if err := ValidatePath(path); err != nil {
		return nil, err
	}
	return git.OpenCached(path)
}
