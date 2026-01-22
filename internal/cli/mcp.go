package cli

import (
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server for LLM integration",
	Long:  `Model Context Protocol server for integration with Claude and other LLMs.`,
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	Long:  `Start the MCP server for LLM tool integration.`,
	RunE:  runMCPServe,
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd)
}

func runMCPServe(cmd *cobra.Command, args []string) error {
	// MCP server implementation will be added in Phase 2
	return RunMCPServer()
}
