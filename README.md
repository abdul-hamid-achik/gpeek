# gpeek

A terminal user interface (TUI) for code review, git visualization, and repository management.

![gpeek screenshot](https://via.placeholder.com/800x500?text=gpeek+screenshot)

## Features

- **Four-panel layout** - Files, Branches, Commits, and Preview panels
- **Syntax-highlighted diffs** - View changes with full syntax highlighting
- **Git operations** - Stage, unstage, commit, push, pull, fetch, and checkout
- **Branch management** - Create, delete, and switch branches
- **Worktree support** - Manage multiple working directories
- **Vim-style keybindings** - Navigate with familiar keys (j/k, g/G, Ctrl+u/d)
- **Theme support** - Built-in themes: Nord, Dracula, Catppuccin Mocha, Gruvbox Dark
- **Configurable** - Customize keybindings, UI options, and behavior

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
# Run in current directory
gpeek

# Run in specific repository
gpeek /path/to/repo
```

## Keybindings

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
| `n` | Create new branch |
| `d` | Delete branch |

### Views & Modals

| Key | Action |
|-----|--------|
| `v` | Toggle diff view modal |
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

## License

MIT License - see [LICENSE](LICENSE) for details.
