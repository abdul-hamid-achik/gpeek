package search

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FilterBar provides inline filtering for panels
type FilterBar struct {
	styles        *ui.Styles
	input         textinput.Model
	active        bool
	width         int

	// Filter options
	caseSensitive bool
	regexMode     bool

	// Filter state
	query         search.Query
	totalCount    int
	matchCount    int
	regexErr      error
}

// NewFilterBar creates a new filter bar component
func NewFilterBar(styles *ui.Styles) *FilterBar {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 100
	ti.SetWidth(30)

	return &FilterBar{
		styles:        styles,
		input:         ti,
		caseSensitive: false,
		regexMode:     false,
	}
}

// Activate shows and focuses the filter bar
func (f *FilterBar) Activate() {
	f.active = true
	f.input.Focus()
}

// Deactivate hides the filter bar and clears the filter
func (f *FilterBar) Deactivate() {
	f.active = false
	f.input.Blur()
	f.input.SetValue("")
	f.query = search.Query{}
}

// Cancel hides the filter bar but keeps the filter text
func (f *FilterBar) Cancel() {
	f.active = false
	f.input.Blur()
}

// Accept accepts the current filter and deactivates
func (f *FilterBar) Accept() {
	f.active = false
	f.input.Blur()
}

// IsActive returns whether the filter bar is active
func (f *FilterBar) IsActive() bool {
	return f.active
}

// SetWidth sets the width of the filter bar
func (f *FilterBar) SetWidth(width int) {
	f.width = width
	targetWidth := width - 20 // Account for decorations
	if targetWidth < 10 {
		targetWidth = 10
	}
	f.input.SetWidth(targetWidth)
}

// SetCounts updates the match/total counts for display
func (f *FilterBar) SetCounts(matchCount, totalCount int) {
	f.matchCount = matchCount
	f.totalCount = totalCount
}

// GetQuery returns the current parsed query
func (f *FilterBar) GetQuery() search.Query {
	return f.query
}

// GetPattern returns the raw filter text
func (f *FilterBar) GetPattern() string {
	return f.input.Value()
}

// HasFilter returns true if there is an active filter
func (f *FilterBar) HasFilter() bool {
	return f.input.Value() != ""
}

// ToggleCaseSensitive toggles case sensitivity
func (f *FilterBar) ToggleCaseSensitive() {
	f.caseSensitive = !f.caseSensitive
	f.updateQuery()
}

// ToggleRegex toggles regex mode
func (f *FilterBar) ToggleRegex() {
	f.regexMode = !f.regexMode
	f.updateQuery()
}

// updateQuery parses the input into a query
func (f *FilterBar) updateQuery() {
	opts := search.QueryOptions{
		DefaultMode:      search.ModeFuzzy,
		DefaultCaseSens:  f.caseSensitive,
		DefaultSmartCase: !f.caseSensitive,
	}

	if f.regexMode {
		opts.DefaultMode = search.ModeRegex
	}

	f.query = search.ParseQuery(f.input.Value(), opts)
	f.regexErr = f.query.RegexError()
}

// HasError returns true if there's a regex compilation error
func (f *FilterBar) HasError() bool {
	return f.regexErr != nil
}

// GetError returns the current regex error message, if any
func (f *FilterBar) GetError() string {
	if f.regexErr != nil {
		return f.regexErr.Error()
	}
	return ""
}

// Update handles input events
func (f *FilterBar) Update(msg tea.Msg) tea.Cmd {
	if !f.active {
		return nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			f.Deactivate()
			return nil
		case "enter":
			f.Accept()
			return nil
		case "alt+c":
			f.ToggleCaseSensitive()
			return nil
		case "alt+r":
			f.ToggleRegex()
			return nil
		}
	}

	f.input, cmd = f.input.Update(msg)
	f.updateQuery()
	return cmd
}

// View renders the filter bar
func (f *FilterBar) View() string {
	if !f.active && !f.HasFilter() {
		return ""
	}

	var parts []string

	// Filter prompt
	prompt := "/ "
	if f.regexMode {
		prompt = "~ "
	}
	parts = append(parts, f.styles.StatusBarKey.Render(prompt))

	// Input field
	if f.active {
		parts = append(parts, f.input.View())
	} else {
		// Show filter value when not active
		val := f.input.Value()
		if len(val) > 20 {
			val = val[:17] + "..."
		}
		parts = append(parts, f.styles.StatusBarValue.Render(val))
	}

	// Show error or match count
	if f.regexErr != nil {
		errStr := " ⚠ " + f.regexErr.Error()
		if len(errStr) > 30 {
			errStr = errStr[:27] + "..."
		}
		parts = append(parts, f.styles.Error.Render(errStr))
	} else {
		countStr := fmt.Sprintf(" (%d/%d)", f.matchCount, f.totalCount)
		parts = append(parts, f.styles.Dim.Render(countStr))
	}

	// Mode indicators
	var indicators []string
	if f.regexMode {
		indicators = append(indicators, f.styles.StatusBarKey.Render("~"))
	}
	if f.caseSensitive {
		indicators = append(indicators, f.styles.StatusBarKey.Render("Aa"))
	} else {
		indicators = append(indicators, f.styles.Dim.Render("Aa"))
	}

	if len(indicators) > 0 {
		parts = append(parts, " ["+strings.Join(indicators, "][")+"]")
	}

	line := strings.Join(parts, "")

	// Create a styled bar
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.styles.Theme.Foreground)).
		Background(lipgloss.Color(f.styles.Theme.Border)).
		Width(f.width).
		Padding(0, 1)

	return barStyle.Render(line)
}

// FilterHeight returns the height of the filter bar (0 if not visible)
func (f *FilterBar) FilterHeight() int {
	if f.active || f.HasFilter() {
		return 1
	}
	return 0
}
