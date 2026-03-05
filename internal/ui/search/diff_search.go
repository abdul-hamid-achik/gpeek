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

// DiffSearchMatch represents a single match in the diff content
type DiffSearchMatch struct {
	Line      int // Line number in the diff content
	StartCol  int // Start column of match
	EndCol    int // End column of match
	LineText  string
}

// DiffSearch provides search functionality for diff content
type DiffSearch struct {
	styles        *ui.Styles
	input         textinput.Model
	active        bool
	width         int

	// Search options
	caseSensitive bool
	regexMode     bool

	// Search state
	pattern      string
	matches      []DiffSearchMatch
	currentMatch int
	content      []string // Lines of diff content
}

// NewDiffSearch creates a new diff search component
func NewDiffSearch(styles *ui.Styles) *DiffSearch {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 100
	ti.SetWidth(30)

	return &DiffSearch{
		styles:        styles,
		input:         ti,
		caseSensitive: false,
		regexMode:     false,
		currentMatch:  -1,
	}
}

// SetContent sets the content to search through
func (d *DiffSearch) SetContent(content string) {
	d.content = strings.Split(content, "\n")
	// Re-run search if we have an active pattern
	if d.pattern != "" {
		d.performSearch()
	}
}

// Activate shows and focuses the search bar
func (d *DiffSearch) Activate() {
	d.active = true
	d.input.Focus()
}

// Deactivate hides the search bar
func (d *DiffSearch) Deactivate() {
	d.active = false
	d.input.Blur()
	d.input.SetValue("")
	d.pattern = ""
	d.matches = nil
	d.currentMatch = -1
}

// IsActive returns whether the search bar is active
func (d *DiffSearch) IsActive() bool {
	return d.active
}

// HasSearch returns true if there's an active search
func (d *DiffSearch) HasSearch() bool {
	return d.pattern != ""
}

// SetWidth sets the width of the search bar
func (d *DiffSearch) SetWidth(width int) {
	d.width = width
	targetWidth := width - 30
	if targetWidth < 10 {
		targetWidth = 10
	}
	d.input.SetWidth(targetWidth)
}

// ToggleCaseSensitive toggles case sensitivity
func (d *DiffSearch) ToggleCaseSensitive() {
	d.caseSensitive = !d.caseSensitive
	d.performSearch()
}

// ToggleRegex toggles regex mode
func (d *DiffSearch) ToggleRegex() {
	d.regexMode = !d.regexMode
	d.performSearch()
}

// NextMatch moves to the next match and returns its line number
func (d *DiffSearch) NextMatch() int {
	if len(d.matches) == 0 {
		return -1
	}
	d.currentMatch++
	if d.currentMatch >= len(d.matches) {
		d.currentMatch = 0
	}
	return d.matches[d.currentMatch].Line
}

// PrevMatch moves to the previous match and returns its line number
func (d *DiffSearch) PrevMatch() int {
	if len(d.matches) == 0 {
		return -1
	}
	d.currentMatch--
	if d.currentMatch < 0 {
		d.currentMatch = len(d.matches) - 1
	}
	return d.matches[d.currentMatch].Line
}

// CurrentMatchLine returns the current match's line number, or -1 if no matches
func (d *DiffSearch) CurrentMatchLine() int {
	if len(d.matches) == 0 || d.currentMatch < 0 {
		return -1
	}
	return d.matches[d.currentMatch].Line
}

// MatchCount returns the number of matches
func (d *DiffSearch) MatchCount() int {
	return len(d.matches)
}

// CurrentMatchIndex returns the current match index (1-based for display)
func (d *DiffSearch) CurrentMatchIndex() int {
	if len(d.matches) == 0 {
		return 0
	}
	return d.currentMatch + 1
}

// GetMatches returns all matches for highlighting
func (d *DiffSearch) GetMatches() []DiffSearchMatch {
	return d.matches
}

// performSearch searches the content for the current pattern
func (d *DiffSearch) performSearch() {
	d.matches = nil
	d.currentMatch = -1

	if d.pattern == "" || len(d.content) == 0 {
		return
	}

	mode := search.ModeExact // Use exact match for diff search
	if d.regexMode {
		mode = search.ModeRegex
	}

	matcher := search.NewMatcher(d.pattern, mode, d.caseSensitive, !d.caseSensitive)

	for i, line := range d.content {
		result := matcher.Match(line)
		if result.Matched {
			for _, m := range result.Matches {
				d.matches = append(d.matches, DiffSearchMatch{
					Line:     i,
					StartCol: m.Start,
					EndCol:   m.End,
					LineText: line,
				})
			}
		}
	}

	if len(d.matches) > 0 {
		d.currentMatch = 0
	}
}

// Update handles input events
func (d *DiffSearch) Update(msg tea.Msg) (tea.Cmd, int) {
	if !d.active {
		return nil, -1
	}

	var cmd tea.Cmd
	scrollTo := -1

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Deactivate()
			return nil, -1
		case "enter":
			// Accept search and stay at current match
			d.active = false
			d.input.Blur()
			return nil, -1
		case "ctrl+n", "tab":
			scrollTo = d.NextMatch()
			return nil, scrollTo
		case "ctrl+p", "shift+tab":
			scrollTo = d.PrevMatch()
			return nil, scrollTo
		case "alt+c":
			d.ToggleCaseSensitive()
			return nil, d.CurrentMatchLine()
		case "alt+r":
			d.ToggleRegex()
			return nil, d.CurrentMatchLine()
		}
	}

	d.input, cmd = d.input.Update(msg)

	// If pattern changed, re-search
	newPattern := d.input.Value()
	if newPattern != d.pattern {
		d.pattern = newPattern
		d.performSearch()
		scrollTo = d.CurrentMatchLine()
	}

	return cmd, scrollTo
}

// View renders the search bar
func (d *DiffSearch) View() string {
	if !d.active && !d.HasSearch() {
		return ""
	}

	var parts []string

	// Search prompt
	prompt := "/ "
	if d.regexMode {
		prompt = "~ "
	}
	parts = append(parts, d.styles.StatusBarKey.Render(prompt))

	// Input field
	if d.active {
		parts = append(parts, d.input.View())
	} else {
		val := d.pattern
		if len(val) > 20 {
			val = val[:17] + "..."
		}
		parts = append(parts, d.styles.StatusBarValue.Render(val))
	}

	// Match count
	var countStr string
	if len(d.matches) == 0 && d.pattern != "" {
		countStr = " (no matches)"
	} else if len(d.matches) > 0 {
		countStr = fmt.Sprintf(" (%d/%d)", d.CurrentMatchIndex(), len(d.matches))
	}
	parts = append(parts, d.styles.Dim.Render(countStr))

	// Mode indicators
	var indicators []string
	if d.regexMode {
		indicators = append(indicators, d.styles.StatusBarKey.Render("~"))
	}
	if d.caseSensitive {
		indicators = append(indicators, d.styles.StatusBarKey.Render("Aa"))
	} else {
		indicators = append(indicators, d.styles.Dim.Render("Aa"))
	}

	if len(indicators) > 0 {
		parts = append(parts, " ["+strings.Join(indicators, "][")+"]")
	}

	// Navigation hint
	if d.active && len(d.matches) > 1 {
		parts = append(parts, d.styles.Dim.Render("  n/N: next/prev"))
	}

	line := strings.Join(parts, "")

	// Create a styled bar
	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.styles.Theme.Foreground)).
		Background(lipgloss.Color(d.styles.Theme.Border)).
		Width(d.width).
		Padding(0, 1)

	return barStyle.Render(line)
}

// IsLineMatch returns true if the given line number has a match
func (d *DiffSearch) IsLineMatch(lineNum int) bool {
	for _, m := range d.matches {
		if m.Line == lineNum {
			return true
		}
	}
	return false
}

// GetLineMatches returns matches for a specific line
func (d *DiffSearch) GetLineMatches(lineNum int) []DiffSearchMatch {
	var lineMatches []DiffSearchMatch
	for _, m := range d.matches {
		if m.Line == lineNum {
			lineMatches = append(lineMatches, m)
		}
	}
	return lineMatches
}
