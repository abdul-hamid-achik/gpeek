# Contributing to gpeek

Thank you for your interest in contributing to gpeek!

## Development Setup

### Prerequisites

- Go 1.21 or later (project uses Go 1.25)
- Git
- A terminal with true color support (for testing themes)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/abdul-hamid-achik/gpeek.git
cd gpeek

# Install dependencies
go mod download

# Build
go build -o gpeek ./cmd/gpeek

# Run
./gpeek
```

### Running Tests

```bash
go test ./...
```

### Linting

```bash
# Install golangci-lint if not already installed
# https://golangci-lint.run/welcome/install/

golangci-lint run
```

## Code Style

- Follow standard Go conventions and formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and reasonably sized

## Project Structure

```
gpeek/
├── cmd/gpeek/          # Application entry point
├── internal/
│   ├── agent/          # AI analyzer (conflict check, change impact, summarization)
│   ├── app/            # Main application logic and keybindings
│   ├── cli/            # Cobra CLI commands
│   ├── config/         # Configuration handling
│   ├── diff/           # Diff parsing, rendering (unified + side-by-side)
│   ├── gh/             # GitHub integration
│   ├── git/            # Git repository operations, commit graph, repo pool
│   ├── mcp/            # MCP server
│   │   └── tools/      # 23 MCP tool handlers
│   ├── memory/         # Noted memory integration
│   ├── search/         # Core search engine (fuzzy, exact, regex)
│   ├── ui/             # Layout, themes, styles
│   │   ├── modals/     # Modal dialogs (diff, blame, stash, help, etc.)
│   │   ├── panels/     # Main panels (files, branches, commits, preview)
│   │   └── search/     # Search UI components (filter bar, diff search)
│   └── version/        # Version info
├── configs/            # Default configuration files
└── themes/             # Built-in theme definitions (YAML)
```

## Making Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests and linter (`go test ./... && golangci-lint run`)
5. Commit your changes (`git commit -m 'Add my feature'`)
6. Push to your fork (`git push origin feature/my-feature`)
7. Open a Pull Request

## Pull Request Guidelines

- Keep PRs focused on a single change
- Include tests for new functionality
- Update documentation if needed
- Ensure all tests pass
- Ensure linter passes with no new warnings

## Reporting Issues

When reporting bugs, please include:

- gpeek version
- Operating system and terminal emulator
- Steps to reproduce
- Expected vs actual behavior
- Any relevant error messages

## Feature Requests

Feature requests are welcome! Please describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
