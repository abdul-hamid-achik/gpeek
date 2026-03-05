package main

import (
	"fmt"
	"os"

	"github.com/abdul-hamid-achik/gpeek/internal/app"
	"github.com/abdul-hamid-achik/gpeek/internal/cli"
	"github.com/abdul-hamid-achik/gpeek/internal/version"
	tea "charm.land/bubbletea/v2"
)

func main() {
	// Check if we're in CLI mode (using subcommands)
	if len(os.Args) > 1 && cli.IsSubcommand(os.Args[1]) {
		if err := cli.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	// Handle legacy flags for TUI mode
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("gpeek", version.Full())
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	// Run TUI mode
	runTUI()
}

func printHelp() {
	fmt.Println("gpeek - TUI for code review, git visualization, and repository management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gpeek [path]              Launch interactive TUI")
	fmt.Println("  gpeek <command> [flags]   Run CLI command (for agents/scripts)")
	fmt.Println()
	fmt.Println("TUI Options:")
	fmt.Println("  --version, -v    Show version")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("CLI Commands (use --format json for structured output):")
	fmt.Println("  status               Show repository status")
	fmt.Println("  diff                 Show changes in working tree or commit")
	fmt.Println("  log                  Show commit history")
	fmt.Println("  summary              Complete repository snapshot (ideal for LLMs)")
	fmt.Println("  blame <file>         Show line-by-line attribution")
	fmt.Println("  branches             List branches")
	fmt.Println("  stashes              List stashes")
	fmt.Println("  tags                 List tags")
	fmt.Println()
	fmt.Println("Semantic Search Commands:")
	fmt.Println("  index commits        Index commits for semantic search")
	fmt.Println("  search commits       Search commits semantically")
	fmt.Println("  similar              Find similar changes")
	fmt.Println()
	fmt.Println("Analysis Commands:")
	fmt.Println("  check-conflicts      Check for merge conflicts (dry-run)")
	fmt.Println("  summarize-changes    Summarize changes between refs")
	fmt.Println("  analyze-changes      Analyze staged/unstaged changes")
	fmt.Println()
	fmt.Println("MCP Server:")
	fmt.Println("  mcp serve            Start MCP server for LLM integration")
	fmt.Println()
	fmt.Println("CLI Global Flags:")
	fmt.Println("  --format, -f     Output format: json, compact, plain (default: plain)")
	fmt.Println("  --quiet, -q      Suppress non-essential output")
	fmt.Println("  --path, -C       Repository path (default: .)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gpeek                               Launch TUI in current directory")
	fmt.Println("  gpeek /path/to/repo                 Launch TUI for specific repo")
	fmt.Println("  gpeek status -f json                Get status as JSON")
	fmt.Println("  gpeek summary -f json               Get complete repo snapshot as JSON")
	fmt.Println("  gpeek diff --staged -f json         Get staged changes as JSON")
	fmt.Println("  gpeek summarize-changes --from main Get changes since main branch")
	fmt.Println()
	fmt.Println("Run 'gpeek <command> --help' for more information on a command.")
}

func runTUI() {
	repoPath := "."
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	model, err := app.New(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
