package panels

import (
	"fmt"
	"strings"
	"time"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type CommitsPanel struct {
	BasePanel
	styles *ui.Styles

	commits []git.Commit
	cursor  int
	offset  int
}

func NewCommitsPanel(styles *ui.Styles) *CommitsPanel {
	return &CommitsPanel{
		styles: styles,
	}
}

func (p *CommitsPanel) SetCommits(commits []git.Commit) {
	p.commits = commits
	if p.cursor >= len(commits) && len(commits) > 0 {
		p.cursor = len(commits) - 1
	}
}

func (p *CommitsPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			p.moveDown()
		case "k", "up":
			p.moveUp()
		case "g":
			p.cursor = 0
			p.offset = 0
		case "G":
			if len(p.commits) > 0 {
				p.cursor = len(p.commits) - 1
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
	if len(p.commits) == 0 {
		return p.styles.Dim.Render("No commits")
	}

	var lines []string

	end := p.offset + p.height
	if end > len(p.commits) {
		end = len(p.commits)
	}

	for i := p.offset; i < end; i++ {
		c := p.commits[i]
		line := p.renderCommit(c, i == p.cursor)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (p *CommitsPanel) renderCommit(c git.Commit, selected bool) string {
	hash := c.Hash[:7]
	msg := c.Message
	if len(msg) > p.width-20 {
		msg = msg[:p.width-23] + "..."
	}

	timeStr := p.formatTime(c.Time)

	graph := p.renderGraph(c)

	line := fmt.Sprintf("%s %s %s %s",
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
	if p.cursor < len(p.commits)-1 {
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
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+p.height {
		p.offset = p.cursor - p.height + 1
	}
}

func (p *CommitsPanel) SelectedCommit() *git.Commit {
	if len(p.commits) == 0 {
		return nil
	}
	if p.cursor < len(p.commits) {
		return &p.commits[p.cursor]
	}
	return nil
}
