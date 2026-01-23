# gpeek

Git visualization for humans and agents. A dual-mode tool providing both an interactive TUI for developers and structured JSON/MCP output for LLMs and automation.

## Features

### For Humans (TUI Mode)
- **Four-panel layout** - Files, Branches, Commits, and Preview panels
- **Syntax-highlighted diffs** - View changes with full syntax highlighting
- **Git operations** - Stage, unstage, commit, push, pull, fetch, and checkout
- **Branch management** - Create, delete, and switch branches
- **Worktree support** - Manage multiple working directories
- **Vim-style keybindings** - Navigate with familiar keys (j/k, g/G, Ctrl+u/d)
- **Theme support** - Built-in themes: Nord, Dracula, Catppuccin Mocha, Gruvbox Dark

### For Agents (CLI + MCP Mode)
- **Structured JSON output** - All commands support `-f json` for machine-readable output
- **MCP server** - Model Context Protocol integration for Claude and other LLMs
- **One-call context** - `gpeek summary` provides complete repo state in a single request
- **Semantic search** - Find commits by intent, not just keywords
- **Smart analysis** - Conflict prediction, change summarization, risk assessment

## Installation

### Using Go

```bash
go install github.com/abdul-hamid-achik/gpeek/cmd/gpeek@latest
```

### Using Homebrew (macOS/Linux)

```bash
brew tap abdul-hamid-achik/tap
brew install gpeek
```

### Download Binary

Download pre-built binaries from the [Releases](https://github.com/abdul-hamid-achik/gpeek/releases) page.

## Quick Start

```bash
# Launch interactive TUI
gpeek

# CLI: Get repo status as JSON
gpeek status -f json

# CLI: Complete repo snapshot for agents
gpeek summary -f json

# Start MCP server for Claude/LLM integration
gpeek mcp serve
```

---

## Agent & LLM Integration

gpeek provides two integration methods for agents: **CLI with JSON output** and **MCP server**.

### CLI Commands

All commands support `--format` (`-f`) with values: `json`, `compact`, `plain`

| Command | Description |
|---------|-------------|
| `gpeek status` | Repository status (staged, unstaged, untracked files) |
| `gpeek diff` | Structured diffs with parsed hunks |
| `gpeek log` | Commit history with filters |
| `gpeek summary` | Complete repo snapshot (recommended for agents) |
| `gpeek blame <file>` | Line-by-line attribution |
| `gpeek branches` | Branch list with tracking info |
| `gpeek stashes` | Stash list with details |
| `gpeek tags` | Tag list with annotations |

#### Global Flags

```
-f, --format string   Output format: json, compact, plain (default "plain")
-C, --path string     Repository path (default ".")
-q, --quiet           Suppress non-essential output
```

#### Example: Get Full Context in One Call

```bash
gpeek summary -f json
```

Response:
```json
{
  "repository": {
    "name": "my-project",
    "path": "/path/to/repo",
    "branch": "main"
  },
  "status": {
    "staged": [],
    "unstaged": [{"path": "src/app.ts", "status": "modified"}],
    "is_clean": false
  },
  "recent_commits": [...],
  "branches": {"current": "main", "local": [...]},
  "stashes": {"count": 0, "entries": []},
  "tags": {"count": 5, "tags": [...]}
}
```

### MCP Server Integration

Start the MCP server for direct integration with Claude or other MCP-compatible LLMs:

```bash
gpeek mcp serve
```

#### Claude Code Configuration

Add to your Claude Code MCP settings (`~/.claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "gpeek": {
      "command": "gpeek",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### Available MCP Tools

| Tool | Description |
|------|-------------|
| `gpeek_status` | Get repository status |
| `gpeek_diff` | Get structured diffs (staged/unstaged/commit) |
| `gpeek_log` | Get commit history with filters |
| `gpeek_summary` | Complete repo snapshot in one call |
| `gpeek_blame` | Line-by-line file attribution |
| `gpeek_branches` | List local/remote branches |
| `gpeek_stashes` | List stashed changes |
| `gpeek_tags` | List repository tags |
| `gpeek_changes_between` | Analyze changes between two refs |
| `gpeek_conflict_check` | Dry-run merge conflict detection |

#### MCP Tool Parameters

**gpeek_summary** (recommended starting point):
```json
{
  "path": ".",        // Repository path (optional)
  "commits": 10       // Number of recent commits (optional)
}
```

**gpeek_diff**:
```json
{
  "path": ".",        // Repository path (optional)
  "file": "",         // Specific file to diff (optional)
  "staged": false,    // Show staged changes (optional)
  "commit": ""        // Show diff for specific commit (optional)
}
```

**gpeek_blame**:
```json
{
  "path": ".",        // Repository path (optional)
  "file": "src/app.ts", // File to blame (required)
  "start_line": 1,    // Start line (optional)
  "end_line": 50      // End line (optional)
}
```

**gpeek_changes_between**:
```json
{
  "path": ".",        // Repository path (optional)
  "from": "v1.0.0",   // Starting ref (required)
  "to": "HEAD"        // Ending ref (optional, default: HEAD)
}
```

**gpeek_conflict_check**:
```json
{
  "path": ".",            // Repository path (optional)
  "branch": "feature/x",  // Branch to check (required)
  "into": "main"          // Target branch (optional, default: current)
}
```

### Smart Analysis Commands

Advanced commands for intelligent code analysis:

```bash
# Check if merging a branch would cause conflicts
gpeek check-conflicts --branch feature/ui -f json

# Summarize changes between two refs
gpeek summarize-changes --from v1.0.0 --to HEAD -f json

# Analyze staged changes for risk assessment
gpeek analyze-changes --staged -f json
```

### Semantic Search

Index and search commits by intent:

```bash
# Index commit history (run once)
gpeek index commits

# Search commits semantically
gpeek search commits "authentication fix" -f json

# Find commits similar to staged changes
gpeek similar --staged -f json
```

---

## TUI Mode

Launch the TUI by running `gpeek` without subcommands:

```bash
gpeek              # Current directory
gpeek /path/to/repo  # Specific repository
```

## TUI Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Ctrl+d` | Page down |
| `Ctrl+u` | Page up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Tab` | Next panel |
| `Shift+Tab` | Previous panel |
| `1` | Focus Files panel |
| `2` | Focus Branches panel |
| `3` | Focus Commits panel |
| `4` | Focus Preview panel |

### Git Operations

| Key | Action |
|-----|--------|
| `s` | Stage file |
| `u` | Unstage file |
| `x` | Discard changes |
| `c` | Commit staged changes |
| `P` | Push to remote |
| `p` | Pull from remote |
| `f` | Fetch from remote |
| `Enter` | Checkout branch (in Branches panel) |
| `Enter` | View commit diff (in Commits panel) |
| `n` | Create new branch |
| `d` | Delete branch |

### Views & Modals

| Key | Action |
|-----|--------|
| `v` | Toggle diff view (Files panel) |
| `w` | Worktree management |
| `?` | Show help |
| `Ctrl+r` | Refresh all panels |
| `q` / `Ctrl+c` | Quit |

## Configuration

gpeek looks for configuration in:
- `$XDG_CONFIG_HOME/gpeek/config.yaml`
- `~/.config/gpeek/config.yaml`

### Example Configuration

```yaml
theme: nord

ui:
  show_icons: true
  date_format: "2006-01-02 15:04"
  relative_dates: true
  confirm_destructive: true
  show_hidden_files: false

keys:
  quit: ["q", "ctrl+c"]
  help: ["?"]
  refresh: ["ctrl+r"]
  stage: ["s"]
  unstage: ["u"]
  commit: ["c"]

git:
  auto_fetch: false
  auto_fetch_interval: 300
  sign_commits: false

github:
  enabled: true
```

## Themes

gpeek ships with several built-in themes:

- **Nord** (default) - Arctic, bluish color scheme
- **Dracula** - Dark theme with vibrant colors
- **Catppuccin Mocha** - Soothing pastel theme
- **Gruvbox Dark** - Retro groove colors

### Custom Themes

Create a theme file in `~/.config/gpeek/themes/`:

```yaml
name: my-theme
background: "#1a1b26"
foreground: "#c0caf5"
primary: "#7aa2f7"
secondary: "#bb9af7"
accent: "#f7768e"
muted: "#565f89"
border: "#3b4261"
selection: "#33467c"
added: "#9ece6a"
removed: "#f7768e"
modified: "#e0af68"
syntax:
  keyword: "#bb9af7"
  string: "#9ece6a"
  number: "#ff9e64"
  comment: "#565f89"
  function: "#7aa2f7"
  type: "#2ac3de"
```

## Requirements

- Git (for repository operations)
- Terminal with 256-color or true color support
- Minimum terminal size: 80x24

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## Why gpeek for Agents?

| Raw Git Commands | gpeek Agent Mode |
|------------------|------------------|
| `git status` - requires text parsing | `gpeek status -f json` - structured data |
| Multiple commands for full context | `gpeek summary -f json` - one call |
| No semantic search | `gpeek search commits "fix login"` |
| Manual conflict checking | `gpeek check-conflicts --branch X` |
| No change analysis | `gpeek analyze-changes --staged` |

## License

MIT License - see [LICENSE](LICENSE) for details.
