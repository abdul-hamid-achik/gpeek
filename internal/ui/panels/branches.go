package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type BranchesPanel struct {
	BasePanel
	styles *ui.Styles

	branches []git.Branch
	current  string
	cursor   int

	worktrees         []git.Worktree
	showWorktrees     bool
	worktreeCursor    int
	inWorktreeSection bool
}

func NewBranchesPanel(styles *ui.Styles) *BranchesPanel {
	return &BranchesPanel{
		styles: styles,
	}
}

func (p *BranchesPanel) SetBranches(branches []git.Branch, current string) {
	p.branches = branches
	p.current = current
	if p.cursor >= len(branches) && len(branches) > 0 {
		p.cursor = len(branches) - 1
	}
}

func (p *BranchesPanel) SetWorktrees(worktrees []git.Worktree) {
	p.worktrees = worktrees
}

func (p *BranchesPanel) Update(msg tea.Msg) tea.Cmd {
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
			p.inWorktreeSection = false
		case "G":
			if p.showWorktrees && len(p.worktrees) > 0 {
				p.inWorktreeSection = true
				p.worktreeCursor = len(p.worktrees) - 1
			} else if len(p.branches) > 0 {
				p.cursor = len(p.branches) - 1
			}
		case "W":
			if len(p.worktrees) > 0 {
				p.showWorktrees = !p.showWorktrees
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

func (p *BranchesPanel) View() string {
	if len(p.branches) == 0 {
		return p.styles.Dim.Render("No branches\n\nPress (n) to create a new branch")
	}

	var lines []string

	header := p.styles.Bold.Render(fmt.Sprintf("Local (%d)", len(p.branches)))
	lines = append(lines, header)

	for i, b := range p.branches {
		line := p.renderBranch(b, i == p.cursor && !p.inWorktreeSection)
		lines = append(lines, line)
	}

	if p.showWorktrees && len(p.worktrees) > 0 {
		lines = append(lines, "")
		header := p.styles.Bold.Render(fmt.Sprintf("Worktrees (%d)", len(p.worktrees)))
		lines = append(lines, header)

		for i, w := range p.worktrees {
			line := p.renderWorktree(w, i == p.worktreeCursor && p.inWorktreeSection)
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")

	if len(lines) > p.height {
		start := 0
		cursorLine := p.getCursorLineIndex()
		if cursorLine > p.height-3 {
			start = cursorLine - p.height + 3
		}
		end := start + p.height
		if end > len(lines) {
			end = len(lines)
			start = end - p.height
			if start < 0 {
				start = 0
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	return content
}

func (p *BranchesPanel) renderBranch(b git.Branch, selected bool) string {
	prefix := "  "
	if selected && p.focused {
		prefix = "> "
	}

	currentMarker := " "
	if b.Name == p.current {
		currentMarker = "*"
	}

	line := prefix + currentMarker + " " + b.Name

	if b.IsRemote {
		line = prefix + currentMarker + " origin/" + b.Name
	}

	if selected && p.focused {
		return p.styles.ListItemSelected.Render(line)
	}

	if b.Name == p.current {
		return p.styles.ListItemActive.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *BranchesPanel) renderWorktree(w git.Worktree, selected bool) string {
	prefix := "  "
	if selected && p.focused {
		prefix = "> "
	}
	line := fmt.Sprintf("%s  %s (%s)", prefix, w.Path, w.Branch)

	if selected && p.focused {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *BranchesPanel) moveDown() {
	if !p.inWorktreeSection {
		if p.cursor < len(p.branches)-1 {
			p.cursor++
		} else if p.showWorktrees && len(p.worktrees) > 0 {
			p.inWorktreeSection = true
			p.worktreeCursor = 0
		}
	} else {
		if p.worktreeCursor < len(p.worktrees)-1 {
			p.worktreeCursor++
		}
	}
}

func (p *BranchesPanel) moveUp() {
	if p.inWorktreeSection {
		if p.worktreeCursor > 0 {
			p.worktreeCursor--
		} else {
			p.inWorktreeSection = false
			p.cursor = len(p.branches) - 1
		}
	} else {
		if p.cursor > 0 {
			p.cursor--
		}
	}
}

func (p *BranchesPanel) getCursorLineIndex() int {
	if !p.inWorktreeSection {
		return 1 + p.cursor
	}
	return 1 + len(p.branches) + 2 + p.worktreeCursor
}

func (p *BranchesPanel) SelectedBranch() *git.Branch {
	if p.inWorktreeSection || len(p.branches) == 0 {
		return nil
	}
	if p.cursor < len(p.branches) {
		return &p.branches[p.cursor]
	}
	return nil
}

func (p *BranchesPanel) SelectedWorktree() *git.Worktree {
	if !p.inWorktreeSection || len(p.worktrees) == 0 {
		return nil
	}
	if p.worktreeCursor < len(p.worktrees) {
		return &p.worktrees[p.worktreeCursor]
	}
	return nil
}
