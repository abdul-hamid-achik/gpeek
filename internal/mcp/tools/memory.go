package tools

import (
	"context"
	"fmt"

	"github.com/abdul-hamid-achik/gpeek/internal/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RememberInput is the input for gpeek_remember
type RememberInput struct {
	Content  string            `json:"content" jsonschema:"description=The content to remember,required"`
	Category string            `json:"category,omitempty" jsonschema:"description=Category for the memory (e.g. 'decision', 'context', 'issue')"`
	Tags     []string          `json:"tags,omitempty" jsonschema:"description=Tags for easy retrieval"`
	Metadata map[string]string `json:"metadata,omitempty" jsonschema:"description=Additional key-value metadata"`
}

// RememberResponse is the response for remember operations
type RememberResponse struct {
	Success bool           `json:"success"`
	Memory  *MemoryInfo    `json:"memory,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// MemoryInfo represents a stored memory
type MemoryInfo struct {
	ID        string            `json:"id,omitempty"`
	Content   string            `json:"content"`
	Category  string            `json:"category"`
	Tags      []string          `json:"tags"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
}

// RegisterRememberTool registers the gpeek_remember tool
func RegisterRememberTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_remember",
		Description: "Store a memory for later recall. Requires 'noted' to be installed (https://github.com/your-org/noted)",
	}, handleRemember)
}

func handleRemember(ctx context.Context, req *mcp.CallToolRequest, input RememberInput) (*mcp.CallToolResult, any, error) {
	if input.Content == "" {
		return ErrorResult("content parameter is required"), nil, nil
	}

	provider := memory.NewNotedProvider()
	if !provider.Available() {
		response := RememberResponse{
			Success: false,
			Error:   "noted is not installed. Install from https://github.com/your-org/noted to use memory features.",
		}
		return ResultJSON(response)
	}

	mem, err := provider.Remember(input.Content, memory.RememberOptions{
		Category: input.Category,
		Tags:     input.Tags,
		Metadata: input.Metadata,
	})
	if err != nil {
		response := RememberResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to store memory: %v", err),
		}
		return ResultJSON(response)
	}

	response := RememberResponse{
		Success: true,
		Memory: &MemoryInfo{
			ID:        mem.ID,
			Content:   mem.Content,
			Category:  mem.Category,
			Tags:      mem.Tags,
			Metadata:  mem.Metadata,
			CreatedAt: mem.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	return ResultJSON(response)
}

// RecallContextInput is the input for gpeek_recall_context
type RecallContextInput struct {
	Path     string   `json:"path,omitempty" jsonschema:"description=Repository path (default: current directory)"`
	Query    string   `json:"query,omitempty" jsonschema:"description=Search query for memories"`
	Category string   `json:"category,omitempty" jsonschema:"description=Filter by category"`
	Tags     []string `json:"tags,omitempty" jsonschema:"description=Filter by tags"`
	Limit    int      `json:"limit,omitempty" jsonschema:"description=Maximum number of memories to return (default: 10)"`
	UseContext bool   `json:"use_context,omitempty" jsonschema:"description=Automatically include context from current branch and recent commits"`
}

// RecallContextResponse is the response for recall operations
type RecallContextResponse struct {
	Success  bool          `json:"success"`
	Memories []MemoryInfo  `json:"memories"`
	Context  *ContextInfo  `json:"context,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ContextInfo contains context used for recall
type ContextInfo struct {
	Branch        string   `json:"branch,omitempty"`
	RecentCommits []string `json:"recent_commits,omitempty"`
}

// RegisterRecallContextTool registers the gpeek_recall_context tool
func RegisterRecallContextTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "gpeek_recall_context",
		Description: "Recall stored memories, optionally using current git context. Requires 'noted' to be installed.",
	}, handleRecallContext)
}

func handleRecallContext(ctx context.Context, req *mcp.CallToolRequest, input RecallContextInput) (*mcp.CallToolResult, any, error) {
	provider := memory.NewNotedProvider()
	if !provider.Available() {
		response := RecallContextResponse{
			Success: false,
			Error:   "noted is not installed. Install from https://github.com/your-org/noted to use memory features.",
		}
		return ResultJSON(response)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	var memories []memory.Memory
	var contextInfo *ContextInfo
	var err error

	if input.UseContext {
		// Get git context
		repo, repoErr := OpenRepo(input.Path)
		if repoErr != nil {
			return ErrorResult(fmt.Sprintf("Failed to open repository: %v", repoErr)), nil, nil
		}

		branch := repo.CurrentBranch()
		commits, _ := repo.Log(5)

		var commitMsgs []string
		for _, c := range commits {
			commitMsgs = append(commitMsgs, c.Message)
		}

		contextInfo = &ContextInfo{
			Branch:        branch,
			RecentCommits: commitMsgs,
		}

		memories, err = provider.RecallContext("", branch, commitMsgs)
	} else {
		memories, err = provider.Recall(memory.RecallOptions{
			Query:    input.Query,
			Category: input.Category,
			Tags:     input.Tags,
			Limit:    limit,
		})
	}

	if err != nil {
		response := RecallContextResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to recall memories: %v", err),
		}
		return ResultJSON(response)
	}

	// Convert to response format
	memoryInfos := make([]MemoryInfo, len(memories))
	for i, m := range memories {
		memoryInfos[i] = MemoryInfo{
			ID:        m.ID,
			Content:   m.Content,
			Category:  m.Category,
			Tags:      m.Tags,
			Metadata:  m.Metadata,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response := RecallContextResponse{
		Success:  true,
		Memories: memoryInfos,
		Context:  contextInfo,
	}

	return ResultJSON(response)
}
