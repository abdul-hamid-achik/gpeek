package ui

import "charm.land/lipgloss/v2"

type Styles struct {
	Theme Theme

	// Base styles
	Base      lipgloss.Style
	Bold      lipgloss.Style
	Dim       lipgloss.Style
	Italic    lipgloss.Style
	Underline lipgloss.Style

	// Panel styles
	Panel           lipgloss.Style
	PanelFocused    lipgloss.Style
	PanelTitle      lipgloss.Style
	PanelTitleFocus lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemActive   lipgloss.Style

	// Status bar
	StatusBar      lipgloss.Style
	StatusBarKey   lipgloss.Style
	StatusBarValue lipgloss.Style
	StatusBarError lipgloss.Style

	// Git status
	Added     lipgloss.Style
	Removed   lipgloss.Style
	Modified  lipgloss.Style
	Renamed   lipgloss.Style
	Untracked lipgloss.Style
	Conflict  lipgloss.Style

	// Diff styles
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffContext lipgloss.Style
	DiffHunk    lipgloss.Style
	DiffMeta    lipgloss.Style
	LineNumber  lipgloss.Style
	SearchMatch lipgloss.Style

	// Modal styles
	Modal       lipgloss.Style
	ModalTitle  lipgloss.Style
	ModalBorder lipgloss.Style

	// Commit graph
	GraphCommit lipgloss.Style
	GraphMerge  lipgloss.Style
	GraphBranch lipgloss.Style

	// Messages
	Error   lipgloss.Style
	Warning lipgloss.Style
	Success lipgloss.Style
	Info    lipgloss.Style

	// Input
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	Placeholder lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Spinner
	Spinner lipgloss.Style
}

func NewStyles(theme Theme) *Styles {
	s := &Styles{Theme: theme}

	// Base styles
	s.Base = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground))

	s.Bold = s.Base.
		Background(lipgloss.Color(theme.Background)).
		Bold(true)

	s.Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted))

	s.Italic = s.Base.Italic(true)

	s.Underline = s.Base.Underline(true)

	// Panel styles
	s.Panel = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Border)).
		Padding(0, 1)

	s.PanelFocused = s.Panel.
		BorderForeground(lipgloss.Color(theme.Primary))

	s.PanelTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted)).
		Bold(true).
		Padding(0, 1)

	s.PanelTitleFocus = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary)).
		Bold(true).
		Padding(0, 1)

	// List styles
	s.ListItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground))

	s.ListItemSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Background)).
		Background(lipgloss.Color(theme.Primary)).
		Bold(true)

	s.ListItemActive = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary)).
		Bold(true)

	// Status bar
	s.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground)).
		Background(lipgloss.Color(theme.Border)).
		Padding(0, 1)

	s.StatusBarKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary)).
		Bold(true)

	s.StatusBarValue = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground))

	s.StatusBarError = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Error)).
		Bold(true)

	// Git status
	s.Added = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Added))

	s.Removed = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Removed))

	s.Modified = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Modified))

	s.Renamed = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Renamed))

	s.Untracked = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Untracked))

	s.Conflict = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Conflict)).
		Bold(true)

	// Diff styles
	s.DiffAdd = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Added))

	s.DiffRemove = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Removed))

	s.DiffContext = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground))

	s.DiffHunk = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary)).
		Bold(true)

	s.DiffMeta = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted))

	s.LineNumber = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Width(4).
		Align(lipgloss.Right)

	s.SearchMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Background)).
		Background(lipgloss.Color(theme.Warning))

	// Modal styles
	s.Modal = lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Background)).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Primary)).
		BorderBackground(lipgloss.Color(theme.Background)).
		Padding(1, 2)

	s.ModalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary)).
		Background(lipgloss.Color(theme.Background)).
		Bold(true).
		Padding(0, 1)

	s.ModalBorder = lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Background)).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Primary)).
		BorderBackground(lipgloss.Color(theme.Background))

	// Commit graph
	s.GraphCommit = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary))

	s.GraphMerge = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Accent))

	s.GraphBranch = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary))

	// Messages
	s.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Error)).
		Bold(true)

	s.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Warning))

	s.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Success))

	s.Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Info))

	// Input
	s.Input = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Foreground)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(theme.Border)).
		Padding(0, 1)

	s.InputFocus = s.Input.
		BorderForeground(lipgloss.Color(theme.Primary))

	s.Placeholder = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted))

	// Help
	s.HelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary)).
		Background(lipgloss.Color(theme.Background)).
		Bold(true)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted)).
		Background(lipgloss.Color(theme.Background))

	// Spinner
	s.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Primary))

	return s
}

var DefaultStyles = NewStyles(NordTheme())
