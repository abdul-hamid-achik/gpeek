package panels

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/search"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	uisearch "github.com/abdul-hamid-achik/gpeek/internal/ui/search"
	tea "github.com/charmbracelet/bubbletea"
)

type BranchesPanel struct {
	BasePanel
	styles *ui.Styles

	allBranches      []git.Branch
	filteredBranches []git.Branch
	current          string
	cursor           int

	worktrees         []git.Worktree
	showWorktrees     bool
	worktreeCursor    int
	inWorktreeSection bool

	tags          []git.Tag
	showTags      bool
	tagCursor     int
	inTagSection  bool

	// Filter support
	filterBar *uisearch.FilterBar
}

func NewBranchesPanel(styles *ui.Styles) *BranchesPanel {
	return &BranchesPanel{
		styles:    styles,
		filterBar: uisearch.NewFilterBar(styles),
	}
}

func (p *BranchesPanel) SetBranches(branches []git.Branch, current string) {
	p.allBranches = branches
	p.current = current
	p.applyFilter()
}

func (p *BranchesPanel) applyFilter() {
	query := p.filterBar.GetQuery()

	if query.Text == "" {
		p.filteredBranches = p.allBranches
	} else {
		p.filteredBranches = search.Filter(p.allBranches, query, func(b git.Branch) string {
			return b.Name
		})
	}

	// Update filter bar counts
	p.filterBar.SetCounts(len(p.filteredBranches), len(p.allBranches))

	// Adjust cursor if needed
	if p.cursor >= len(p.filteredBranches) && len(p.filteredBranches) > 0 {
		p.cursor = len(p.filteredBranches) - 1
	}
	if len(p.filteredBranches) == 0 {
		p.cursor = 0
	}
}

func (p *BranchesPanel) SetWorktrees(worktrees []git.Worktree) {
	p.worktrees = worktrees
}

func (p *BranchesPanel) SetTags(tags []git.Tag) {
	p.tags = tags
}

func (p *BranchesPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)
	p.filterBar.SetWidth(width)
}

func (p *BranchesPanel) Update(msg tea.Msg) tea.Cmd {
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
			p.inWorktreeSection = false
		case "G":
			if p.showTags && len(p.tags) > 0 {
				p.inWorktreeSection = false
				p.inTagSection = true
				p.tagCursor = len(p.tags) - 1
			} else if p.showWorktrees && len(p.worktrees) > 0 {
				p.inTagSection = false
				p.inWorktreeSection = true
				p.worktreeCursor = len(p.worktrees) - 1
			} else if len(p.filteredBranches) > 0 {
				p.inTagSection = false
				p.inWorktreeSection = false
				p.cursor = len(p.filteredBranches) - 1
			}
		case "W":
			if len(p.worktrees) > 0 {
				p.showWorktrees = !p.showWorktrees
			}
		case "t":
			if len(p.tags) > 0 {
				p.showTags = !p.showTags
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
	branches := p.filteredBranches

	if len(branches) == 0 && len(p.allBranches) == 0 {
		return p.styles.Dim.Render("No branches\n\nPress (n) to create a new branch")
	}

	if len(branches) == 0 && p.filterBar.HasFilter() {
		content := p.styles.Dim.Render("No matching branches")
		if p.filterBar.IsActive() || p.filterBar.HasFilter() {
			content += "\n" + p.filterBar.View()
		}
		return content
	}

	var lines []string

	header := p.styles.Bold.Render(fmt.Sprintf("Local (%d)", len(branches)))
	if p.filterBar.HasFilter() {
		header = p.styles.Bold.Render(fmt.Sprintf("Local (%d/%d)", len(branches), len(p.allBranches)))
	}
	lines = append(lines, header)

	for i, b := range branches {
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

	if p.showTags && len(p.tags) > 0 {
		lines = append(lines, "")
		header := p.styles.Bold.Render(fmt.Sprintf("Tags (%d)", len(p.tags)))
		lines = append(lines, header)

		for i, t := range p.tags {
			line := p.renderTag(t, i == p.tagCursor && p.inTagSection)
			lines = append(lines, line)
		}
	}

	// Calculate available height for content (reserve space for filter bar)
	contentHeight := p.height
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		contentHeight -= p.filterBar.FilterHeight()
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	content := strings.Join(lines, "\n")

	if len(lines) > contentHeight {
		start := 0
		cursorLine := p.getCursorLineIndex()
		if cursorLine > contentHeight-3 {
			start = cursorLine - contentHeight + 3
		}
		end := start + contentHeight
		if end > len(lines) {
			end = len(lines)
			start = end - contentHeight
			if start < 0 {
				start = 0
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	// Add filter bar at bottom if active
	if p.filterBar.IsActive() || p.filterBar.HasFilter() {
		content += "\n" + p.filterBar.View()
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

func (p *BranchesPanel) renderTag(t git.Tag, selected bool) string {
	prefix := "  "
	if selected && p.focused {
		prefix = "> "
	}

	icon := "○"
	if t.IsAnnotated {
		icon = "●"
	}

	line := fmt.Sprintf("%s %s %s", prefix, icon, t.Name)
	if t.Hash != "" {
		hash := t.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		line += fmt.Sprintf(" (%s)", hash)
	}

	if selected && p.focused {
		return p.styles.ListItemSelected.Render(line)
	}

	return p.styles.ListItem.Render(line)
}

func (p *BranchesPanel) moveDown() {
	if p.inTagSection {
		// In tags section
		if p.tagCursor < len(p.tags)-1 {
			p.tagCursor++
		}
	} else if p.inWorktreeSection {
		// In worktrees section
		if p.worktreeCursor < len(p.worktrees)-1 {
			p.worktreeCursor++
		} else if p.showTags && len(p.tags) > 0 {
			p.inWorktreeSection = false
			p.inTagSection = true
			p.tagCursor = 0
		}
	} else {
		// In branches section
		if p.cursor < len(p.filteredBranches)-1 {
			p.cursor++
		} else if p.showWorktrees && len(p.worktrees) > 0 {
			p.inWorktreeSection = true
			p.worktreeCursor = 0
		} else if p.showTags && len(p.tags) > 0 {
			p.inTagSection = true
			p.tagCursor = 0
		}
	}
}

func (p *BranchesPanel) moveUp() {
	if p.inTagSection {
		// In tags section
		if p.tagCursor > 0 {
			p.tagCursor--
		} else if p.showWorktrees && len(p.worktrees) > 0 {
			p.inTagSection = false
			p.inWorktreeSection = true
			p.worktreeCursor = len(p.worktrees) - 1
		} else {
			p.inTagSection = false
			if len(p.filteredBranches) > 0 {
				p.cursor = len(p.filteredBranches) - 1
			}
		}
	} else if p.inWorktreeSection {
		// In worktrees section
		if p.worktreeCursor > 0 {
			p.worktreeCursor--
		} else {
			p.inWorktreeSection = false
			if len(p.filteredBranches) > 0 {
				p.cursor = len(p.filteredBranches) - 1
			}
		}
	} else {
		// In branches section
		if p.cursor > 0 {
			p.cursor--
		}
	}
}

func (p *BranchesPanel) getCursorLineIndex() int {
	if p.inTagSection {
		idx := 1 + len(p.filteredBranches)
		if p.showWorktrees && len(p.worktrees) > 0 {
			idx += 2 + len(p.worktrees)
		}
		return idx + 2 + p.tagCursor
	}
	if p.inWorktreeSection {
		return 1 + len(p.filteredBranches) + 2 + p.worktreeCursor
	}
	return 1 + p.cursor
}

func (p *BranchesPanel) SelectedBranch() *git.Branch {
	if p.inWorktreeSection || p.inTagSection || len(p.filteredBranches) == 0 {
		return nil
	}
	if p.cursor < len(p.filteredBranches) {
		return &p.filteredBranches[p.cursor]
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

func (p *BranchesPanel) SelectedTag() *git.Tag {
	if !p.inTagSection || len(p.tags) == 0 {
		return nil
	}
	if p.tagCursor < len(p.tags) {
		return &p.tags[p.tagCursor]
	}
	return nil
}

func (p *BranchesPanel) InTagSection() bool {
	return p.inTagSection
}

// IsFiltering returns true if the filter bar is active
func (p *BranchesPanel) IsFiltering() bool {
	return p.filterBar.IsActive()
}

// ClearFilter clears the current filter
func (p *BranchesPanel) ClearFilter() {
	p.filterBar.Deactivate()
	p.applyFilter()
}
