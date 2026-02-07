package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConflictCheckInput is the input for gpeek_conflict_check
type ConflictCheckInput struct {
	Path   string `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Branch string `json:"branch" jsonschema:"description=Branch to check for conflicts,required"`
	Into   string `json:"into,omitempty" jsonschema:"description=Target branch to merge into (default: current branch)"`
}

// ConflictCheckResponse is the response for conflict check
type ConflictCheckResponse struct {
	Branch         string `json:"branch"`
	Into           string `json:"into"`
	WouldConflict  bool   `json:"would_conflict"`
	SafeToMerge    bool   `json:"safe_to_merge"`
	Recommendation string `json:"recommendation"`
	Note           string `json:"note"`
}

// RegisterConflictCheckTool registers the gpeek_conflict_check tool
func RegisterConflictCheckTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_conflict_check",
		Description: "Check if merging a branch would cause conflicts (dry-run)",
	}, handleConflictCheck)
}

func handleConflictCheck(ctx context.Context, req *mcp.CallToolRequest, input ConflictCheckInput) (*mcp.CallToolResult, any, error) {
	repo, err := OpenRepo(input.Path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open repository: %v", err)), nil, nil
	}

	if input.Branch == "" {
		return ErrorResult("branch parameter is required"), nil, nil
	}

	into := input.Into
	if into == "" {
		into = repo.CurrentBranch()
	}

	// Use git merge-tree to check for conflicts without modifying the working tree
	repoPath := DefaultPath(input.Path)
	cmd := exec.Command("git", "merge-tree", "--write-tree", "--no-messages", into, input.Branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()

	wouldConflict := false
	conflictFiles := []string{}

	if err != nil {
		// Non-zero exit means conflicts were found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			wouldConflict = true
			// Parse conflict file names from output
			for _, line := range strings.Split(string(output), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, " ") {
					// Skip the tree hash on first line
					continue
				}
				if strings.HasPrefix(line, "CONFLICT") {
					conflictFiles = append(conflictFiles, line)
				}
			}
		} else {
			// git merge-tree --write-tree may not be available (requires Git 2.38+)
			// Fall back to merge-base approach
			return fallbackConflictCheck(repoPath, input.Branch, into)
		}
	}

	response := ConflictCheckResponse{
		Branch:        input.Branch,
		Into:          into,
		WouldConflict: wouldConflict,
		SafeToMerge:   !wouldConflict,
	}

	if wouldConflict {
		response.Recommendation = fmt.Sprintf("Resolve conflicts before merging %s into %s", input.Branch, into)
		response.Note = fmt.Sprintf("Found %d conflict(s): %s", len(conflictFiles), strings.Join(conflictFiles, "; "))
	} else {
		response.Recommendation = fmt.Sprintf("Safe to merge %s into %s", input.Branch, into)
		response.Note = "No conflicts detected via git merge-tree"
	}

	return ResultJSON(response)
}

// fallbackConflictCheck uses merge-base for older Git versions (< 2.38)
func fallbackConflictCheck(repoPath, branch, into string) (*mcp.CallToolResult, any, error) {
	// Get merge base
	baseCmd := exec.Command("git", "merge-base", into, branch)
	baseCmd.Dir = repoPath
	baseOutput, err := baseCmd.Output()
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to find merge base: %v", err)), nil, nil
	}
	mergeBase := strings.TrimSpace(string(baseOutput))

	// Use old-style merge-tree with three args
	cmd := exec.Command("git", "merge-tree", mergeBase, into, branch)
	cmd.Dir = repoPath
	output, _ := cmd.Output()

	// If output contains conflict markers, there are conflicts
	wouldConflict := strings.Contains(string(output), "changed in both")

	response := ConflictCheckResponse{
		Branch:        branch,
		Into:          into,
		WouldConflict: wouldConflict,
		SafeToMerge:   !wouldConflict,
	}

	if wouldConflict {
		response.Recommendation = fmt.Sprintf("Resolve conflicts before merging %s into %s", branch, into)
		response.Note = "Conflicts detected via merge-tree (legacy mode)"
	} else {
		response.Recommendation = fmt.Sprintf("Safe to merge %s into %s", branch, into)
		response.Note = "No conflicts detected via merge-tree (legacy mode)"
	}

	return ResultJSON(response)
}
