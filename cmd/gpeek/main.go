package main

import (
	"fmt"
	"os"

	"github.com/abdul-hamid-achik/gpeek/internal/app"
	"github.com/abdul-hamid-achik/gpeek/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("gpeek", version.Full())
			return
		case "--help", "-h", "help":
			fmt.Println("gpeek - TUI for code review, git visualization, and repository management")
			fmt.Println()
			fmt.Println("Usage: gpeek [path]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --version, -v    Show version")
			fmt.Println("  --help, -h       Show this help")
			fmt.Println()
			fmt.Println("Run 'gpeek' in a git repository to start the TUI.")
			return
		}
	}

	repoPath := "."
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	model, err := app.New(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
