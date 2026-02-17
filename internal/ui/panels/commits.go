package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CommitsPanel struct {
	BasePanel
	styles *ui.Styles

	allCommits      []git.Commit
	filteredCommits []git.Commit
	cursor          int
	offset          int

	// Filter support
	filterBar *uisearch.FilterBar
}

func NewCommitsPanel(styles *ui.Styles) *CommitsPanel {
	return &CommitsPanel{
		styles:    styles,
		filterBar: uisearch.NewFilterBar(styles),
	}
}

func (p *CommitsPanel) SetCommits(commits []git.Commit) {
	p.allCommits = commits
	p.applyFilter()
}

func (p *CommitsPanel) applyFilter() {
	query := p.filterBar.GetQuery()

	if query.Text == "" {
		p.filteredCommits = p.allCommits
	} else {
		// Filter by message or author
		p.filteredCommits = search.Filter(p.allCommits, query, func(c git.Commit) string {
			return c.Message + " " + c.Author
		})
	}

	// Update filter bar counts
	p.filterBar.SetCounts(len(p.filteredCommits), len(p.allCommits))

	// Adjust cursor if needed
	if p.cursor >= len(p.filteredCommits) && len(p.filteredCommits) > 0 {
		p.cursor = len(p.filteredCommits) - 1
	}
	if len(p.filteredCommits) == 0 {
		p.cursor = 0
	}
	p.adjustOffset()
}

func (p *CommitsPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)
	p.filterBar.SetWidth(width)
}

func (p *CommitsPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	// If filter bar is active, handle its input first
	if p.filterBar.IsActive() {
		cmd := p.filterBar.Update(msg)
		p.applyFilter()
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			p.filterBar.Activate()
			return nil
		case "esc":
			if p.filterBar.HasFilter() {
				p.filterBar.Deactivate()
				p.applyFilter()
				return nil
			}
		case "j", "down":
			p.moveDown()
		case "k", "up":
			p.moveUp()
		case "g":
			p.cursor = 0
			p.offset = 0
		case "G":
			if len(p.filteredCommits) > 0 {
				p.cursor = len(p.filteredCommits) - 1
				p.adjustOffset()
			}
		case "ctrl+d":
			for i := 0; i < p.height/2; i++ {
				p.moveDown()
			}
		case "ctrl+u":
			for i := 0; i < p.height/2; i++ {
				p.moveUp()
			}
		}
	}

	return nil
}

func (p *CommitsPanel) View() string {
	commits := p.filteredCommits

	if len(commits) == 0 && len(p.allCommits) == 0 {
		return p.styles.Dim.Render("No commits\n\nMake your first commit with (c)")
	}

	if len(commits) == 0 && p.filterBar.HasFilter() {
		content := p.styles.Dim.Render("No matching commits")
		if p.filterBar.IsActive() || p.filterBar.HasFilter() {
			content += "\n" + p.filterBar.View()
		}
		return content
	}

	// Calculate available height for commits (reserve space for filter bar)
	contentHeight := p.height
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		contentHeight -= p.filterBar.FilterHeight()
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	var lines []string

	end := p.offset + contentHeight
	if end > len(commits) {
		end = len(commits)
	}

	for i := p.offset; i < end; i++ {
		c := commits[i]
		line := p.renderCommit(c, i == p.cursor)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	// Add filter bar at bottom if active
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		content += "\n" + p.filterBar.View()
	}

	return content
}

func (p *CommitsPanel) renderCommit(c git.Commit, selected bool) string {
	hash := c.Hash
	if len(hash) > 7 {
		hash = hash[:7]
	}
	msg := c.Message
	timeStr := p.formatTime(c.Time)
	graph := p.renderGraph(c)

	selPrefix := "  "
	if selected && p.focused {
		selPrefix = "> "
	}

	// Calculate available width for message
	prefix := selPrefix + graph + " " + hash + " "
	suffix := " " + timeStr
	prefixWidth := lipgloss.Width(prefix)
	suffixWidth := lipgloss.Width(suffix)
	availableWidth := p.width - prefixWidth - suffixWidth - 2
	if availableWidth < 0 {
		availableWidth = 0
	}

	// Truncate message if needed using visual width
	if lipgloss.Width(msg) > availableWidth && availableWidth > 3 {
		runes := []rune(msg)
		for i := len(runes); i > 0; i-- {
			truncated := string(runes[:i]) + "..."
			if lipgloss.Width(truncated) <= availableWidth {
				msg = truncated
				break
			}
		}
	}

	line := fmt.Sprintf("%s%s %s %s %s",
		selPrefix,
		graph,
		p.styles.GraphCommit.Render(hash),
		msg,
		p.styles.Dim.Render(timeStr),
	)

	if selected && p.focused {
		return p.styles.ListItemSelected.Render(line)
	}

	return line
}

func (p *CommitsPanel) renderGraph(c git.Commit) string {
	if c.IsMerge {
		return p.styles.GraphMerge.Render("●")
	}
	return p.styles.GraphCommit.Render("○")
}

func (p *CommitsPanel) formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		return fmt.Sprintf("%dw ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		return fmt.Sprintf("%dy ago", years)
	}
}

func (p *CommitsPanel) moveDown() {
	if p.cursor < len(p.filteredCommits)-1 {
		p.cursor++
		p.adjustOffset()
	}
}

func (p *CommitsPanel) moveUp() {
	if p.cursor > 0 {
		p.cursor--
		p.adjustOffset()
	}
}

func (p *CommitsPanel) adjustOffset() {
	// Account for filter bar height
	contentHeight := p.height
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		contentHeight -= p.filterBar.FilterHeight()
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+contentHeight {
		p.offset = p.cursor - contentHeight + 1
	}
}

func (p *CommitsPanel) SelectedCommit() *git.Commit {
	if len(p.filteredCommits) == 0 {
		return nil
	}
	if p.cursor < len(p.filteredCommits) {
		return &p.filteredCommits[p.cursor]
	}
	return nil
}

// IsFiltering returns true if the filter bar is active
func (p *CommitsPanel) IsFiltering() bool {
	return p.filterBar.IsActive()
}

// ClearFilter clears the current filter
func (p *CommitsPanel) ClearFilter() {
	p.filterBar.Deactivate()
	p.applyFilter()
}
