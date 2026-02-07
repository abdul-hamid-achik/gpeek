package modals

import (
	"sort"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents an action in the command palette
type Command struct {
	ID          string
	Title       string
	Description string
	Category    string
	Keybinding  string
	Action      func() tea.Cmd
}

// PaletteModal is a command palette modal
type PaletteModal struct {
	BaseModal
	input       textinput.Model
	commands    []Command
	filtered    []Command
	selected    int
	onExecute   func(Command) tea.Cmd
	styles      PaletteStyles
}

// PaletteStyles holds styles for the palette
type PaletteStyles struct {
	Container       lipgloss.Style
	Input           lipgloss.Style
	Item            lipgloss.Style
	ItemSelected    lipgloss.Style
	ItemTitle       lipgloss.Style
	ItemDescription lipgloss.Style
	ItemKeybinding  lipgloss.Style
	Category        lipgloss.Style
}

// DefaultPaletteStyles returns default styles using the given theme
func DefaultPaletteStyles(styles *ui.Styles) PaletteStyles {
	theme := styles.Theme
	return PaletteStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Primary)).
			Padding(0, 1),
		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(0, 1),
		Item: lipgloss.NewStyle().
			Padding(0, 1),
		ItemSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color(theme.Selection)).
			Foreground(lipgloss.Color(theme.Foreground)),
		ItemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Foreground)),
		ItemDescription: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)),
		ItemKeybinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Accent)).
			Bold(true),
		Category: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Italic(true),
	}
}

// NewPaletteModal creates a new command palette modal
func NewPaletteModal(commands []Command, onExecute func(Command) tea.Cmd, width, height int, uiStyles ...*ui.Styles) *PaletteModal {
	input := textinput.New()
	input.Placeholder = "Type to search commands..."
	input.Focus()
	input.CharLimit = 100
	input.Width = width - 4

	var ps PaletteStyles
	if len(uiStyles) > 0 && uiStyles[0] != nil {
		ps = DefaultPaletteStyles(uiStyles[0])
	} else {
		ps = defaultFallbackPaletteStyles()
	}

	m := &PaletteModal{
		input:     input,
		commands:  commands,
		filtered:  commands,
		selected:  0,
		onExecute: onExecute,
		styles:    ps,
	}
	m.width = width
	m.height = height

	return m
}

// defaultFallbackPaletteStyles returns hardcoded fallback styles when no theme is provided
func defaultFallbackPaletteStyles() PaletteStyles {
	return PaletteStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")).
			Padding(0, 1),
		Item: lipgloss.NewStyle().
			Padding(0, 1),
		ItemSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")),
		ItemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")),
		ItemDescription: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		ItemKeybinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true),
		Category: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
	}
}

// Update handles input
func (m *PaletteModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if m.ShouldClose() {
		return nil, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.Close()
			return nil, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if len(m.filtered) > 0 && m.selected < len(m.filtered) {
				cmd := m.filtered[m.selected]
				m.Close()
				if m.onExecute != nil {
					return nil, m.onExecute(cmd)
				}
				if cmd.Action != nil {
					return nil, cmd.Action()
				}
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "ctrl+p"))):
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "ctrl+n"))):
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+u"))):
			m.selected = 0
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))):
			m.selected = len(m.filtered) - 1
			if m.selected < 0 {
				m.selected = 0
			}
			return m, nil
		}
	}

	// Update input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Filter commands based on input
	m.filterCommands()

	return m, cmd
}

// filterCommands filters the command list using fuzzy matching (consistent with panel filter bars)
func (m *PaletteModal) filterCommands() {
	input := strings.TrimSpace(m.input.Value())
	if input == "" {
		m.filtered = m.commands
		m.selected = 0
		return
	}

	query := search.ParseQuery(input, search.DefaultQueryOptions())

	// Use FilterWithScore to rank results by relevance
	scored := search.FilterWithScore(m.commands, query, func(cmd Command) string {
		return cmd.Title + " " + cmd.Description + " " + cmd.Category
	})

	// Sort by score descending (best matches first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	m.filtered = make([]Command, len(scored))
	for i, s := range scored {
		m.filtered[i] = s.Item
	}

	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
		if m.selected < 0 {
			m.selected = 0
		}
	}
}

// View renders the palette
func (m *PaletteModal) View() string {
	var b strings.Builder

	// Input field
	inputView := m.styles.Input.Render(m.input.View())
	b.WriteString(inputView)
	b.WriteString("\n\n")

	// Commands list
	maxVisible := m.height - 6 // Account for input and borders
	if maxVisible < 1 {
		maxVisible = 5
	}

	start := 0
	if m.selected >= maxVisible {
		start = m.selected - maxVisible + 1
	}

	currentCategory := ""
	visibleCount := 0

	for i := start; i < len(m.filtered) && visibleCount < maxVisible; i++ {
		cmd := m.filtered[i]

		// Category header
		if cmd.Category != currentCategory {
			currentCategory = cmd.Category
			categoryView := m.styles.Category.Render("  " + currentCategory)
			b.WriteString(categoryView)
			b.WriteString("\n")
		}

		// Command item
		var itemStyle lipgloss.Style
		if i == m.selected {
			itemStyle = m.styles.ItemSelected
		} else {
			itemStyle = m.styles.Item
		}

		titlePart := m.styles.ItemTitle.Render(cmd.Title)
		if cmd.Keybinding != "" {
			keyPart := m.styles.ItemKeybinding.Render(" [" + cmd.Keybinding + "]")
			titlePart += keyPart
		}

		descPart := ""
		if cmd.Description != "" {
			descPart = "\n  " + m.styles.ItemDescription.Render(cmd.Description)
		}

		itemView := itemStyle.Width(m.width - 4).Render(titlePart + descPart)
		b.WriteString(itemView)
		b.WriteString("\n")
		visibleCount++
	}

	if len(m.filtered) == 0 {
		noResults := m.styles.ItemDescription.Render("  No commands found")
		b.WriteString(noResults)
	}

	return m.styles.Container.Width(m.width).Height(m.height).Render(b.String())
}

// DefaultCommands returns a standard set of commands
func DefaultCommands() []Command {
	return []Command{
		// Navigation
		{ID: "focus_files", Title: "Focus Files Panel", Category: "Navigation", Keybinding: "1"},
		{ID: "focus_branches", Title: "Focus Branches Panel", Category: "Navigation", Keybinding: "2"},
		{ID: "focus_commits", Title: "Focus Commits Panel", Category: "Navigation", Keybinding: "3"},
		{ID: "focus_preview", Title: "Focus Preview Panel", Category: "Navigation", Keybinding: "4"},

		// Git Operations
		{ID: "stage", Title: "Stage File", Description: "Stage the selected file", Category: "Git", Keybinding: "s"},
		{ID: "unstage", Title: "Unstage File", Description: "Unstage the selected file", Category: "Git", Keybinding: "u"},
		{ID: "discard", Title: "Discard Changes", Description: "Discard changes in file", Category: "Git", Keybinding: "x"},
		{ID: "commit", Title: "Commit", Description: "Create a new commit", Category: "Git", Keybinding: "c"},
		{ID: "push", Title: "Push", Description: "Push to remote", Category: "Git", Keybinding: "P"},
		{ID: "pull", Title: "Pull", Description: "Pull from remote", Category: "Git", Keybinding: "p"},
		{ID: "fetch", Title: "Fetch", Description: "Fetch from remote", Category: "Git", Keybinding: "f"},

		// Branch Operations
		{ID: "checkout", Title: "Checkout Branch", Description: "Switch to selected branch", Category: "Branch", Keybinding: "enter"},
		{ID: "new_branch", Title: "New Branch", Description: "Create a new branch", Category: "Branch", Keybinding: "n"},
		{ID: "delete_branch", Title: "Delete Branch", Description: "Delete selected branch", Category: "Branch", Keybinding: "d"},

		// Stash & Worktree
		{ID: "stash", Title: "Stash", Description: "Manage stashes", Category: "Stash", Keybinding: "S"},
		{ID: "worktree", Title: "Worktree", Description: "Manage worktrees", Category: "Worktree", Keybinding: "w"},

		// View
		{ID: "diff_mode", Title: "Toggle Diff View", Description: "Switch diff viewing mode", Category: "View", Keybinding: "v"},
		{ID: "blame", Title: "Blame", Description: "Show file blame", Category: "View", Keybinding: "b"},
		{ID: "show_commit", Title: "Show Commit", Description: "Show commit details", Category: "View", Keybinding: "enter"},

		// Search
		{ID: "filter", Title: "Filter Panel", Description: "Filter current panel", Category: "Search", Keybinding: "/"},
		{ID: "search", Title: "Global Search", Description: "Search across repository", Category: "Search", Keybinding: "ctrl+p"},

		// System
		{ID: "refresh", Title: "Refresh", Description: "Refresh repository state", Category: "System", Keybinding: "ctrl+r"},
		{ID: "git_config", Title: "Git Config", Description: "Edit git configuration", Category: "System", Keybinding: "C"},
		{ID: "help", Title: "Help", Description: "Show help", Category: "System", Keybinding: "?"},
		{ID: "quit", Title: "Quit", Description: "Exit gpeek", Category: "System", Keybinding: "q"},
	}
}
