# Changelog

All notable changes to gpeek will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- **Merge branch** (`m` in Branches panel) -- Merge the selected branch into the current branch. Shows a confirmation dialog before executing. Errors (including conflicts) are surfaced in the status bar.
- **Rebase** (`r` in Branches panel) -- Rebase the current branch onto the selected branch. Confirmation required. Also available via the command palette.
- **Cherry-pick** (`e` in Commits panel) -- Apply the selected commit to the current branch. Confirmation required.
- **Soft reset** (`R` in Commits panel) -- Move HEAD to the selected commit while keeping all changes staged. Safe — no work is lost.
- **Hard reset** (command palette only) -- Move HEAD to the selected commit and discard all uncommitted changes. Shows an explicit destruction warning. Not bound to a key to prevent accidental use.
- **Revert commit** (`y` in Commits panel) -- Create a new commit that undoes the selected commit. Non-destructive; safe on shared branches.
- **Responsive commit modal** -- The commit message textarea now scales with terminal width (30–56 chars) instead of a fixed 60-char box.
- **Diff header truncation** -- Long file paths in diff headers are now truncated with `...` to fit the available panel or modal width. Hunk headers (e.g., long function signatures after `@@`) are truncated as well.
- **Stage/unstage status messages** -- Staging or unstaging a single file now shows a `"Staged path/to/file"` / `"Unstaged path/to/file"` confirmation in the status bar (bulk operations already showed this; single-file was missing).
- **Commit graph visualization** -- The Commits panel now renders ASCII-art branch/merge lines instead of simple dots. Graph is computed once on data load and cached for performance.
- **Remote branches** -- The Branches panel shows remote branches in a new "Remote" section. Press `R` to toggle visibility. Remote branches render with a dimmer style.
- **Side-by-side diff view** -- Press `S` in the Diff modal to toggle between unified and side-by-side diff views. The current mode is shown in the modal header.
- **Panel zoom** -- Press `z` to maximize the focused panel to full screen. Press `z` again to restore the 2x2 grid layout. Also available via the command palette.
- **Auto-expand single-file diffs** -- When the Preview panel loads a diff with exactly one file, it automatically expands instead of requiring manual navigation.
- **Scroll position indicators** -- Panel titles now show cursor position (e.g., "Files (3/12)", "Commits (47/238)").
- **Persistent `ctrl+k` hint** -- The command palette shortcut is always visible in the status bar regardless of which panel is focused.
- **Panel accessor methods** -- `CursorPosition()` and `TotalItems()` public methods on Files, Branches, and Commits panels.

### Fixed

- **Commit modal blank line** -- The commit modal body no longer shows a phantom blank line between the staged files list and the message area when amend mode is off.

### Fixed

- **YAML theme files** -- All four built-in YAML themes (Nord, Catppuccin Mocha, Gruvbox Dark, Dracula) now include the `subtle` field and use correct `muted` values matching the Go-defined defaults. Previously, loading themes from YAML produced invisible text due to insufficient contrast (e.g., Nord `muted` was 1.6:1 contrast ratio).
- **Search highlight readability** -- The `SearchMatch` style now sets an explicit foreground color (theme background) so highlighted matches are always readable against the warning-colored background.

### Changed

- **Regex compilation** -- `stashPattern` in `stash.go` and `emailRegex` in `config.go` are now compiled once at package level instead of on every function call.
- **Sorting algorithm** -- Replaced O(n^2) bubble sort with `sort.Slice` (O(n log n)) in `analysis.go` for hot files and language detection.
- **Language map allocation** -- The `extensionToLanguage` map in `analysis.go` is now a package-level variable instead of being recreated on every call.

### Removed

- Custom `max()` functions in `layout.go` and `diff/word.go` replaced by Go builtin (available since Go 1.21).

## [0.12.0] - 2026-02-17

### Added

- 7 MCP write tools: `gpeek_stage`, `gpeek_unstage`, `gpeek_commit`, `gpeek_stash_save`, `gpeek_stash_pop`, `gpeek_stash_drop`, `gpeek_discard` (23 total)
- Input modal responsive width (adapts to terminal size)
- Stash modal shows `[tool]` tag for `.claude`/`.vecgrep` stash entries
- Preview panel shows empty state message
- Small terminal guards on all panels (min 1x1)
- ~100 new tests: search 34% to 80%, diff 36% to 52%

### Fixed

- Critical crash cluster: empty diff bounds, unguarded focusedFile access
- Stash cursor -1 on empty list, blame visibleLines guard
- Various nil safety and hash truncation hardening

## [0.11.1] - 2026-02-07

### Fixed

- Discard/Unstage data loss prevention
- MCP path traversal vulnerability
- Search integer conversion safety
- Preview panel `G` key panic on empty content
- Iterator early exit, palette execution, stash drop confirm
- Help resize, confirm y/Y handling, branches navigation
- Diff deduplication via shared `content.go`

## [0.11.0] - 2026-02-01

### Added

- Command palette (`Ctrl+K`) with fuzzy search
- Agent analyzer: conflict check, change impact, change summarization
- Memory integration via noted
- Semantic code search via vecgrep

## [0.10.0] - 2026-01-15

### Added

- MCP server with 16 read-only tools
- Structured JSON output for all CLI commands
- Blame modal, stash modal, worktree modal
- Git config modal
- Search modal with branch/commit/worktree tabs

## [0.9.0] - 2025-12-01

### Added

- Initial release
- Four-panel TUI layout (Files, Branches, Commits, Preview)
- Vim-style keybindings
- Syntax-highlighted diffs
- 4 built-in dark themes (Nord, Catppuccin Mocha, Gruvbox Dark, Dracula)
- Custom YAML theme support
- Basic git operations (stage, unstage, commit, push, pull, fetch, checkout)
