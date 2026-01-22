package cli

import (
	"os"

	"github.com/abdul-hamid-achik/gpeek/internal/version"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	formatFlag string
	quietFlag  bool
	pathFlag   string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gpeek",
	Short: "Git visualization for humans and agents",
	Long: `gpeek - TUI for code review, git visualization, and repository management

When invoked without subcommands, gpeek launches the interactive TUI.
Use subcommands for CLI access with structured output (ideal for LLMs/agents).`,
	Version: version.Full(),
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "plain", "Output format: json, compact, plain")
	RootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-essential output")
	RootCmd.PersistentFlags().StringVarP(&pathFlag, "path", "C", ".", "Repository path")

	// Add all subcommands
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(diffCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(summaryCmd)
	RootCmd.AddCommand(blameCmd)
	RootCmd.AddCommand(branchesCmd)
	RootCmd.AddCommand(stashesCmd)
	RootCmd.AddCommand(tagsCmd)
	RootCmd.AddCommand(mcpCmd)
}

// IsSubcommand checks if the given argument is a known subcommand
func IsSubcommand(arg string) bool {
	commands := []string{
		"status", "diff", "log", "summary", "blame",
		"branches", "stashes", "tags", "mcp",
		"search", "index", "similar",
		"check-conflicts", "summarize-changes", "analyze-changes",
		"help", "completion",
	}
	for _, cmd := range commands {
		if arg == cmd {
			return true
		}
	}
	return false
}

// Execute runs the CLI
func Execute() error {
	return RootCmd.Execute()
}

// GetFormat returns the output format
func GetFormat() string {
	return formatFlag
}

// GetPath returns the repository path
func GetPath() string {
	if pathFlag == "" {
		return "."
	}
	return pathFlag
}

// IsQuiet returns whether quiet mode is enabled
func IsQuiet() bool {
	return quietFlag
}

// ExitWithError prints an error and exits
func ExitWithError(err error) {
	if formatFlag == "json" {
		outputJSON(map[string]interface{}{
			"error": map[string]interface{}{
				"message": err.Error(),
			},
		})
	} else {
		_, _ = os.Stderr.WriteString("Error: " + err.Error() + "\n")
	}
	os.Exit(1)
}
