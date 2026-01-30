package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchChangesInput is the input for gpeek_search_changes
type SearchChangesInput struct {
	Path  string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Query string `json:"query" jsonschema:"description=Search query (text to find in commit messages or file changes),required"`
	Since string `json:"since,omitempty" jsonschema:"description=Search commits since date (e.g. '1 week ago' or '2024-01-01')"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default: 20)"`
}

// SearchChangesResponse is the response for search changes
type SearchChangesResponse struct {
	Query   string               `json:"query"`
	Results []SearchChangeResult `json:"results"`
	Total   int                  `json:"total"`
}

// SearchChangeResult represents a single search result
type SearchChangeResult struct {
	Commit      CommitInfo `json:"commit"`
	MatchedIn   string     `json:"matched_in"` // "message", "file_path", or "diff"
	MatchedText string     `json:"matched_text,omitempty"`
	Score       float64    `json:"score,omitempty"`
}

// RegisterSearchChangesTool registers the gpeek_search_changes tool
func RegisterSearchChangesTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_search_changes",
		Description: "Search for text in commit messages, file paths, or diffs within the repository history",
	}, handleSearchChanges)
}

func handleSearchChanges(ctx context.Context, req *mcp.CallToolRequest, input SearchChangesInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Query == "" {
		return ErrorResult("query parameter is required"), nil, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	// Search in recent commits
	commits, err := repo.Log(200) // Search in last 200 commits
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get commits: %v", err)), nil, nil
	}

	queryLower := strings.ToLower(input.Query)
	var results []SearchChangeResult

	for _, c := range commits {
		if len(results) >= limit {
			break
		}

		// Search in commit message
		if strings.Contains(strings.ToLower(c.Message), queryLower) {
			results = append(results, createSearchResult(c, "message", c.Message, 1.0))
			continue
		}

		// Search in diff content
		diff, err := repo.CommitDiff(c.Hash)
		if err != nil {
			continue
		}

		// Check for matches in file paths
		if strings.Contains(strings.ToLower(diff), queryLower) {
			// Extract the matching context
			matchedText := extractMatchContext(diff, input.Query)
			results = append(results, createSearchResult(c, "diff", matchedText, 0.8))
		}
	}

	response := SearchChangesResponse{
		Query:   input.Query,
		Results: results,
		Total:   len(results),
	}

	return ResultJSON(response)
}

func createSearchResult(c git.Commit, matchedIn, matchedText string, score float64) SearchChangeResult {
	shortHash := c.Hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}
	return SearchChangeResult{
		Commit: CommitInfo{
			Hash:      c.Hash,
			ShortHash: shortHash,
			Message:   c.Message,
			Author:    c.Author,
			Email:     c.Email,
			Time:      c.Time,
			TimeAgo:   TimeAgo(c.Time),
			IsMerge:   c.IsMerge,
		},
		MatchedIn:   matchedIn,
		MatchedText: matchedText,
		Score:       score,
	}
}

// extractMatchContext extracts context around a match
func extractMatchContext(text, query string) string {
	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	idx := strings.Index(textLower, queryLower)
	if idx == -1 {
		return ""
	}

	// Find line boundaries
	lineStart := strings.LastIndex(text[:idx], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}
	lineEnd := strings.Index(text[idx:], "\n")
	if lineEnd == -1 {
		lineEnd = len(text)
	} else {
		lineEnd = idx + lineEnd
	}

	result := text[lineStart:lineEnd]
	if len(result) > 200 {
		result = result[:200] + "..."
	}
	return strings.TrimSpace(result)
}
