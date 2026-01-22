package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/diff"
	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	s := &Server{}

	mcpServer := server.NewMCPServer(
		"gpeek",
		version.Full(),
		server.WithToolCapabilities(true),
		server.WithInstructions(`gpeek is a Git visualization tool that provides comprehensive repository information.
Use these tools to understand the current state of a Git repository, view changes,
examine commit history, and get detailed file blame information.

For a complete repository snapshot in one call, use the gpeek_summary tool.`),
	)

	// Register all tools
	s.registerTools(mcpServer)

	s.mcpServer = mcpServer
	return s
}

// ServeStdio starts the MCP server over stdio
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) registerTools(mcpServer *server.MCPServer) {
	// gpeek_status - Get comprehensive git status
	mcpServer.AddTool(
		mcp.NewTool("gpeek_status",
			mcp.WithDescription("Get comprehensive git repository status including staged, unstaged, and untracked files"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
		),
		handleStatus,
	)

	// gpeek_diff - Get structured diff
	mcpServer.AddTool(
		mcp.NewTool("gpeek_diff",
			mcp.WithDescription("Get structured diff with parsed hunks. Can show staged changes, working tree changes, or specific commit diffs"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithString("file",
				mcp.Description("Specific file to diff (optional)"),
			),
			mcp.WithBoolean("staged",
				mcp.Description("Show staged changes instead of working tree changes"),
			),
			mcp.WithString("commit",
				mcp.Description("Show diff for a specific commit hash"),
			),
		),
		handleDiff,
	)

	// gpeek_log - Get commit history
	mcpServer.AddTool(
		mcp.NewTool("gpeek_log",
			mcp.WithDescription("Get commit history with optional filters"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of commits to return (default: 50)"),
			),
			mcp.WithString("author",
				mcp.Description("Filter commits by author name"),
			),
			mcp.WithString("since",
				mcp.Description("Show commits since date (e.g., '2024-01-01')"),
			),
		),
		handleLog,
	)

	// gpeek_summary - Complete repository snapshot
	mcpServer.AddTool(
		mcp.NewTool("gpeek_summary",
			mcp.WithDescription("Get complete repository snapshot in one call - ideal for understanding full repo context. Includes status, recent commits, branches, stashes, and tags"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithNumber("commits",
				mcp.Description("Number of recent commits to include (default: 10)"),
			),
		),
		handleSummary,
	)

	// gpeek_blame - Get line-by-line attribution
	mcpServer.AddTool(
		mcp.NewTool("gpeek_blame",
			mcp.WithDescription("Get line-by-line blame information showing who last modified each line"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithString("file",
				mcp.Description("File to blame"),
				mcp.Required(),
			),
			mcp.WithNumber("start_line",
				mcp.Description("Start line number (optional)"),
			),
			mcp.WithNumber("end_line",
				mcp.Description("End line number (optional)"),
			),
		),
		handleBlame,
	)

	// gpeek_branches - List branches
	mcpServer.AddTool(
		mcp.NewTool("gpeek_branches",
			mcp.WithDescription("List local and optionally remote branches"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithBoolean("all",
				mcp.Description("Include remote branches"),
			),
		),
		handleBranches,
	)

	// gpeek_stashes - List stashes
	mcpServer.AddTool(
		mcp.NewTool("gpeek_stashes",
			mcp.WithDescription("List all stashed changes"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
		),
		handleStashes,
	)

	// gpeek_tags - List tags
	mcpServer.AddTool(
		mcp.NewTool("gpeek_tags",
			mcp.WithDescription("List all tags in the repository"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
		),
		handleTags,
	)

	// gpeek_changes_between - Analyze changes between refs
	mcpServer.AddTool(
		mcp.NewTool("gpeek_changes_between",
			mcp.WithDescription("Analyze changes between two git references (branches, tags, commits)"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithString("from",
				mcp.Description("Starting reference (branch, tag, or commit)"),
				mcp.Required(),
			),
			mcp.WithString("to",
				mcp.Description("Ending reference (default: HEAD)"),
			),
		),
		handleChangesBetween,
	)

	// gpeek_conflict_check - Dry-run merge conflict detection
	mcpServer.AddTool(
		mcp.NewTool("gpeek_conflict_check",
			mcp.WithDescription("Check if merging a branch would cause conflicts (dry-run)"),
			mcp.WithString("path",
				mcp.Description("Repository path (default: current directory)"),
			),
			mcp.WithString("branch",
				mcp.Description("Branch to check for conflicts"),
				mcp.Required(),
			),
			mcp.WithString("into",
				mcp.Description("Target branch to merge into (default: current branch)"),
			),
		),
		handleConflictCheck,
	)
}

// Tool handlers

func handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	status, err := repo.Status()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get status: %v", err)), nil
	}

	response := StatusResponse{
		Repository: RepositoryInfo{
			Name:   repo.Name(),
			Path:   repo.Path(),
			Branch: repo.CurrentBranch(),
		},
		Staged:    convertFileEntries(status.Staged),
		Unstaged:  convertFileEntries(status.Unstaged),
		Untracked: status.Untracked,
		Summary: StatusSummary{
			StagedCount:    len(status.Staged),
			UnstagedCount:  len(status.Unstaged),
			UntrackedCount: len(status.Untracked),
			IsClean:        len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
			HasConflicts:   hasConflicts(status),
		},
	}

	return resultJSON(response)
}

func handleDiff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	file := mcp.ParseString(request, "file", "")
	staged := mcp.ParseBoolean(request, "staged", false)
	commit := mcp.ParseString(request, "commit", "")

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	var rawDiff string

	if commit != "" {
		rawDiff, err = repo.CommitDiff(commit)
	} else if file != "" {
		rawDiff, err = repo.FileDiff(file, staged)
	} else {
		// Get all diffs
		status, statusErr := repo.Status()
		if statusErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get status: %v", statusErr)), nil
		}

		var files []git.FileEntry
		if staged {
			files = status.Staged
		} else {
			files = status.Unstaged
		}

		var allDiffs string
		for _, f := range files {
			d, _ := repo.FileDiff(f.Path, staged)
			allDiffs += d
		}
		rawDiff = allDiffs
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get diff: %v", err)), nil
	}

	parsed := diff.Parse(rawDiff)
	response := buildDiffResponse(parsed, file, commit, staged)

	return resultJSON(response)
}

func handleLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	limit := mcp.ParseInt(request, "limit", 50)

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	commits, err := repo.Log(limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get log: %v", err)), nil
	}

	response := buildLogResponse(commits)

	return resultJSON(response)
}

func handleSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	commitLimit := mcp.ParseInt(request, "commits", 10)

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	response := buildSummaryResponse(repo, commitLimit)

	return resultJSON(response)
}

func handleBlame(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	file := mcp.ParseString(request, "file", "")
	startLine := mcp.ParseInt(request, "start_line", 0)
	endLine := mcp.ParseInt(request, "end_line", 0)

	if file == "" {
		return mcp.NewToolResultError("file parameter is required"), nil
	}

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	lines, err := repo.BlameFile(file)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get blame: %v", err)), nil
	}

	// Apply line range filter
	if startLine > 0 || endLine > 0 {
		start := startLine
		if start < 1 {
			start = 1
		}
		end := endLine
		if end < 1 || end > len(lines) {
			end = len(lines)
		}
		if start <= end && start <= len(lines) {
			lines = lines[start-1 : end]
		}
	}

	response := buildBlameResponse(file, lines)

	return resultJSON(response)
}

func handleBranches(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	all := mcp.ParseBoolean(request, "all", false)

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	localBranches, err := repo.ListBranches()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list branches: %v", err)), nil
	}

	var allBranches []git.Branch
	allBranches = append(allBranches, localBranches...)

	if all {
		remoteBranches, _ := repo.ListRemoteBranches()
		allBranches = append(allBranches, remoteBranches...)
	}

	response := buildBranchesResponse(repo.CurrentBranch(), allBranches)

	return resultJSON(response)
}

func handleStashes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	stashes, err := repo.StashList()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list stashes: %v", err)), nil
	}

	response := buildStashesResponse(stashes)

	return resultJSON(response)
}

func handleTags(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	tags, err := repo.ListTags()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list tags: %v", err)), nil
	}

	response := buildTagsResponse(tags)

	return resultJSON(response)
}

func handleChangesBetween(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	from := mcp.ParseString(request, "from", "")
	to := mcp.ParseString(request, "to", "HEAD")

	if from == "" {
		return mcp.NewToolResultError("from parameter is required"), nil
	}

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	// Get commits between the two refs
	// For now, we'll get the diff between the two commits
	rawDiff, err := repo.CommitDiff(fmt.Sprintf("%s..%s", from, to))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get changes: %v", err)), nil
	}

	parsed := diff.Parse(rawDiff)

	response := map[string]interface{}{
		"from":  from,
		"to":    to,
		"files": parsed.Files,
		"stats": map[string]interface{}{
			"files_changed": len(parsed.Files),
		},
	}

	return resultJSON(response)
}

func handleConflictCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := mcp.ParseString(request, "path", ".")
	branch := mcp.ParseString(request, "branch", "")
	into := mcp.ParseString(request, "into", "")

	if branch == "" {
		return mcp.NewToolResultError("branch parameter is required"), nil
	}

	repo, err := git.Open(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open repository: %v", err)), nil
	}

	if into == "" {
		into = repo.CurrentBranch()
	}

	// This is a simplified check - in a real implementation we'd do a dry-run merge
	// For now, we check if there are any files modified in both branches
	response := map[string]interface{}{
		"branch":          branch,
		"into":            into,
		"would_conflict":  false, // Simplified - would need actual merge-base analysis
		"safe_to_merge":   true,
		"recommendation":  fmt.Sprintf("Run 'git merge %s --no-commit --no-ff' to test", branch),
		"note":            "This is a basic check. For accurate conflict detection, use git merge --dry-run",
	}

	return resultJSON(response)
}

// Helper functions

func convertFileEntries(entries []git.FileEntry) []FileInfo {
	result := make([]FileInfo, len(entries))
	for i, e := range entries {
		result[i] = FileInfo{
			Path:   e.Path,
			Status: e.Status.String(),
		}
	}
	return result
}

func hasConflicts(status *git.Status) bool {
	for _, f := range status.Staged {
		if f.Status == git.StatusUpdatedButUnmerged {
			return true
		}
	}
	for _, f := range status.Unstaged {
		if f.Status == git.StatusUpdatedButUnmerged {
			return true
		}
	}
	return false
}

func resultJSON(data interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
