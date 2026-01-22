package cli

import (
	"github.com/abdul-hamid-achik/gpeek/internal/mcp"
)

// RunMCPServer starts the MCP server
func RunMCPServer() error {
	server := mcp.NewServer()
	return server.ServeStdio()
}
