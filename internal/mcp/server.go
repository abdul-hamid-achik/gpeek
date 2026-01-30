package mcp

import (
	"context"

	"github.com/abdul-hamid-achik/gpeek/internal/mcp/tools"
	"github.com/abdul-hamid-achik/gpeek/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server
type Server struct {
	mcpServer *mcp.Server
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	s := &Server{}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "gpeek",
		Version: version.Full(),
	}, &mcp.ServerOptions{
		Instructions: `gpeek is a Git visualization tool that provides comprehensive repository information.
Use these tools to understand the current state of a Git repository, view changes,
examine commit history, and get detailed file blame information.

For a complete repository snapshot in one call, use the gpeek_summary tool.`,
	})

	// Register all tools
	tools.RegisterAllTools(mcpServer)

	s.mcpServer = mcpServer
	return s
}

// ServeStdio starts the MCP server over stdio
func (s *Server) ServeStdio() error {
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}
