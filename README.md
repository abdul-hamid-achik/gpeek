# gpeek

Git visualization for humans and agents. A dual-mode tool providing both an interactive TUI for developers and structured JSON/MCP output for LLMs and automation.

## Features

### For Humans (TUI Mode)
- **Four-panel layout** - Files, Branches, Commits, and Preview panels
- **Syntax-highlighted diffs** - View changes with full syntax highlighting
- **Git operations** - Stage, unstage, commit, push, pull, fetch, and checkout
- **Branch management** - Create, delete, and switch branches
- **Worktree support** - Manage multiple working directories
- **Command palette** - Quick access to all commands via `Ctrl+K`
- **Vim-style keybindings** - Navigate with familiar keys (j/k, g/G, Ctrl+u/d)
- **Theme support** - Built-in themes: Nord, Dracula, Catppuccin Mocha, Gruvbox Dark

### For Agents (CLI + MCP Mode)
- **Structured JSON output** - All commands support `-f json` for machine-readable output
- **MCP server** - Model Context Protocol integration for Claude and other LLMs
- **16 MCP tools** - Comprehensive git operations, search, and analysis
- **One-call context** - `gpeek summary` provides complete repo state in a single request
- **Semantic search** - Find commits by intent, not just keywords
- **Smart analysis** - Conflict prediction, change summarization, risk assessment
- **Memory integration** - Store and recall context across sessions (requires [noted](https://github.com/your-org/noted))

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

### Optional Dependencies

- **vecgrep** - For semantic code search (graceful fallback to text search when unavailable)
- **noted** - For memory features (`gpeek_remember`, `gpeek_recall_context`)

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
| `gpeek stashes` | List stashed changes |
| `gpeek tags` | List repository tags |

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
  "tags": {"count": 5, "tags": [...]},
  "enhanced": {
    "hot_files": [{"path": "main.go", "change_count": 15, "authors": ["dev1"]}],
    "languages": [{"name": "Go", "file_count": 50, "percentage": 80.5}],
    "project_type": "Go",
    "suggestions": ["Review frequently changed files", "Consider adding tests"]
  }
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

gpeek provides 16 MCP tools organized by category:

**Core Tools**
| Tool | Description |
|------|-------------|
| `gpeek_status` | Get repository status |
| `gpeek_diff` | Get structured diffs (staged/unstaged/commit) with optional blame |
| `gpeek_log` | Get commit history with filters and pagination |
| `gpeek_summary` | Complete repo snapshot with enhanced analysis |
| `gpeek_blame` | Line-by-line file attribution |
| `gpeek_branches` | List local/remote branches |
| `gpeek_stashes` | List stashed changes |
| `gpeek_tags` | List repository tags |
| `gpeek_changes_between` | Analyze changes between two refs |
| `gpeek_conflict_check` | Dry-run merge conflict detection |

**Agent Tools**
| Tool | Description |
|------|-------------|
| `gpeek_file_history` | Get commit history for a specific file |
| `gpeek_code_owners` | Identify code owners (CODEOWNERS or blame-based) |

**Search Tools**
| Tool | Description |
|------|-------------|
| `gpeek_search_changes` | Search commits by message, author, or content |
| `gpeek_change_impact` | Analyze impact of staged/unstaged changes |

**Memory Tools** (requires [noted](https://github.com/your-org/noted))
| Tool | Description |
|------|-------------|
| `gpeek_remember` | Store context for later recall |
| `gpeek_recall_context` | Retrieve stored memories with git context |

#### MCP Tool Parameters

**gpeek_summary** (recommended starting point):
```json
{
  "path": ".",        // Repository path (optional)
  "commits": 10,      // Number of recent commits (optional)
  "enhanced": true    // Include hot files, languages, suggestions (optional)
}
```

**gpeek_diff**:
```json
{
  "path": ".",           // Repository path (optional)
  "file": "",            // Specific file to diff (optional)
  "staged": false,       // Show staged changes (optional)
  "commit": "",          // Show diff for specific commit (optional)
  "include_blame": false,// Include blame info in diff (optional)
  "context_lines": 3     // Number of context lines (optional)
}
```

**gpeek_file_history**:
```json
{
  "path": ".",           // Repository path (optional)
  "file": "src/app.ts",  // File to get history for (required)
  "limit": 50,           // Maximum commits to return (optional)
  "offset": 0            // Skip first N commits for pagination (optional)
}
```

**gpeek_code_owners**:
```json
{
  "path": ".",           // Repository path (optional)
  "files": ["src/"],     // Files/directories to check (optional)
  "use_blame": true      // Fall back to blame if no CODEOWNERS (optional)
}
```

**gpeek_search_changes**:
```json
{
  "path": ".",           // Repository path (optional)
  "query": "auth fix",   // Search query (required)
  "author": "",          // Filter by author (optional)
  "since": "1 week ago", // Filter by date (optional)
  "limit": 20            // Maximum results (optional)
}
```

**gpeek_change_impact**:
```json
{
  "path": ".",           // Repository path (optional)
  "staged": true,        // Analyze staged changes (optional)
  "files": []            // Specific files to analyze (optional)
}
```

**gpeek_blame**:
```json
{
  "path": ".",           // Repository path (optional)
  "file": "src/app.ts",  // File to blame (required)
  "start_line": 1,       // Start line (optional)
  "end_line": 50         // End line (optional)
}
```

**gpeek_changes_between**:
```json
{
  "path": ".",           // Repository path (optional)
  "from": "v1.0.0",      // Starting ref (required)
  "to": "HEAD"           // Ending ref (optional, default: HEAD)
}
```

**gpeek_conflict_check**:
```json
{
  "path": ".",           // Repository path (optional)
  "branch": "feature/x", // Branch to check (required)
  "into": "main"         // Target branch (optional, default: current)
}
```

**gpeek_remember** (requires noted):
```json
{
  "content": "Decision: use PostgreSQL for user data",
  "category": "decision",
  "tags": ["database", "architecture"],
  "metadata": {"ticket": "PROJ-123"}
}
```

**gpeek_recall_context** (requires noted):
```json
{
  "query": "database",   // Search query (optional)
  "category": "decision",// Filter by category (optional)
  "tags": ["architecture"],// Filter by tags (optional)
  "use_context": true,   // Use current git context for recall (optional)
  "limit": 10            // Maximum memories to return (optional)
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
| `Ctrl+k` | Open command palette |
| `v` | Toggle diff view (Files panel) |
| `w` | Worktree management |
| `/` | Search in current panel |
| `?` | Show help |
| `Ctrl+r` | Refresh all panels |
| `q` / `Ctrl+c` | Quit |

### Command Palette

Press `Ctrl+K` to open the command palette for quick access to all commands:

- Type to filter commands
- Use `↑`/`↓` or `j`/`k` to navigate
- Press `Enter` to execute
- Press `Escape` to close

Available commands include git operations, navigation, search, and system actions.

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
  command_palette: ["ctrl+k"]

git:
  auto_fetch: false
  auto_fetch_interval: 300
  sign_commits: false

github:
  enabled: true

mcp:
  cache_enabled: true
  cache_max_repos: 10
  vecgrep_enabled: true
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

## Architecture

### Repository Caching

gpeek uses an LRU cache for repository objects to improve MCP server performance:

- Repositories are cached for 5 minutes by default
- Maximum 10 repositories in cache (configurable)
- Automatic cleanup of stale entries

### Search Provider Interface

Search functionality uses a provider interface for graceful degradation:

1. **VecgrepProvider** - Semantic search when vecgrep is installed
2. **FallbackProvider** - Text-based search as fallback

## Requirements

- Git (for repository operations)
- Terminal with 256-color or true color support
- Minimum terminal size: 80x24

### Optional

- **vecgrep** - For enhanced semantic search
- **noted** - For memory/context persistence

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
| No memory across sessions | `gpeek_remember` / `gpeek_recall_context` |

## License

MIT License - see [LICENSE](LICENSE) for details.
