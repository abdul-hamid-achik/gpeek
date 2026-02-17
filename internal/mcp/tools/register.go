package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAllTools registers all gpeek MCP tools
func RegisterAllTools(server *mcp.Server) {
	// Core tools
	RegisterStatusTool(server)
	RegisterDiffTool(server)
	RegisterLogTool(server)
	RegisterSummaryTool(server)
	RegisterBlameTool(server)
	RegisterBranchesTool(server)
	RegisterStashesTool(server)
	RegisterTagsTool(server)
	RegisterChangesBetweenTool(server)
	RegisterConflictCheckTool(server)

	// Agent tools (Phase 2)
	RegisterFileHistoryTool(server)
	RegisterCodeOwnersTool(server)

	// Search tools (Phase 3)
	RegisterSearchChangesTool(server)
	RegisterChangeImpactTool(server)

	// Memory tools (Phase 5) - requires noted
	RegisterRememberTool(server)
	RegisterRecallContextTool(server)

	// Write tools (Phase 6)
	RegisterStageTool(server)
	RegisterUnstageTool(server)
	RegisterCommitWriteTool(server)
	RegisterStashSaveTool(server)
	RegisterStashPopTool(server)
	RegisterStashDropTool(server)
	RegisterDiscardTool(server)
}
