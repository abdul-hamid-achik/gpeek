package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	Refresh    key.Binding
	FocusNext  key.Binding
	FocusPrev  key.Binding
	FocusFiles key.Binding
	FocusBranches key.Binding
	FocusCommits  key.Binding
	FocusPreview  key.Binding
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Select     key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
	Stage      key.Binding
	Unstage    key.Binding
	Discard    key.Binding
	Commit     key.Binding
	Push       key.Binding
	Pull       key.Binding
	Fetch      key.Binding
	Checkout   key.Binding
	NewBranch  key.Binding
	DeleteBranch key.Binding
	DiffMode   key.Binding
	Worktree   key.Binding
	ShowCommit key.Binding
	// Search bindings
	FilterPanel  key.Binding
	GlobalSearch key.Binding
	SearchNext   key.Binding
	SearchPrev   key.Binding
	GitConfig    key.Binding
	// New feature bindings
	Stash          key.Binding
	Blame          key.Binding
	CommandPalette key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev panel"),
		),
		FocusFiles: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "files"),
		),
		FocusBranches: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "branches"),
		),
		FocusCommits: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "commits"),
		),
		FocusPreview: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "preview"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle select"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Stage: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stage"),
		),
		Unstage: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "unstage"),
		),
		Discard: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "discard"),
		),
		Commit: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "commit"),
		),
		Push: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "push"),
		),
		Pull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull"),
		),
		Fetch: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fetch"),
		),
		Checkout: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "checkout"),
		),
		NewBranch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new branch"),
		),
		DeleteBranch: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete branch"),
		),
		DiffMode: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle diff view"),
		),
		Worktree: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "worktree"),
		),
		ShowCommit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show commit"),
		),
		FilterPanel: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		GlobalSearch: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "search"),
		),
		SearchNext: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		SearchPrev: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev match"),
		),
		GitConfig: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "git config"),
		),
		Stash: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stash"),
		),
		Blame: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "blame"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "command palette"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.FocusNext, k.Up, k.Down, k.Select}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom},
		{k.FocusNext, k.FocusPrev, k.FocusFiles, k.FocusBranches, k.FocusCommits, k.FocusPreview},
		{k.Stage, k.Unstage, k.Discard, k.Commit},
		{k.Push, k.Pull, k.Fetch},
		{k.Checkout, k.NewBranch, k.DeleteBranch},
		{k.DiffMode, k.ShowCommit, k.Worktree, k.Stash, k.Blame},
		{k.FilterPanel, k.GlobalSearch, k.SearchNext, k.SearchPrev},
		{k.GitConfig, k.Refresh, k.Help, k.Quit},
		{k.CommandPalette},
	}
}
