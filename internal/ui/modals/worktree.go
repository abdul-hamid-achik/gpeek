package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorktreeMode int

const (
	WorktreeModeList WorktreeMode = iota
	WorktreeModeCreate
	WorktreeModeConfirmDelete
)

type WorktreeModal struct {
	BaseModal
	styles    *ui.Styles
	mode      WorktreeMode
	worktrees []git.Worktree
	cursor    int
	repo      *git.Repository

	pathInput    textinput.Model
	branchInput  textinput.Model
	focusedInput int
	err          string

	// Confirmation state
	confirmPath    string
	confirmFocused int // 0 = Yes, 1 = No
}

func NewWorktreeModal(styles *ui.Styles, worktrees []git.Worktree, repo *git.Repository) *WorktreeModal {
	pathInput := textinput.New()
	pathInput.Placeholder = "Path for new worktree"
	pathInput.Width = 40

	branchInput := textinput.New()
	branchInput.Placeholder = "Branch name (optional)"
	branchInput.Width = 40

	return &WorktreeModal{
		styles:      styles,
		worktrees:   worktrees,
		repo:        repo,
		pathInput:   pathInput,
		branchInput: branchInput,
	}
}

func (m *WorktreeModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation mode separately
		if m.mode == WorktreeModeConfirmDelete {
			switch msg.String() {
			case "esc", "n", "N":
				m.mode = WorktreeModeList
				m.confirmPath = ""
				return m, nil
			case "y", "Y":
				if err := m.repo.RemoveWorktree(m.confirmPath); err != nil {
					m.err = err.Error()
				} else {
					m.worktrees, _ = m.repo.ListWorktrees()
					if m.cursor >= len(m.worktrees) && m.cursor > 0 {
						m.cursor--
					}
				}
				m.mode = WorktreeModeList
				m.confirmPath = ""
				return m, nil
			case "enter":
				if m.confirmFocused == 0 {
					// Yes selected
					if err := m.repo.RemoveWorktree(m.confirmPath); err != nil {
						m.err = err.Error()
					} else {
						m.worktrees, _ = m.repo.ListWorktrees()
						if m.cursor >= len(m.worktrees) && m.cursor > 0 {
							m.cursor--
						}
					}
				}
				m.mode = WorktreeModeList
				m.confirmPath = ""
				return m, nil
			case "tab", "left", "right", "h", "l":
				m.confirmFocused = 1 - m.confirmFocused
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "esc":
			if m.mode == WorktreeModeCreate {
				m.mode = WorktreeModeList
				m.err = ""
				return m, nil
			}
			return nil, nil

		case "q":
			if m.mode == WorktreeModeList {
				return nil, nil
			}

		case "n":
			if m.mode == WorktreeModeList {
				m.mode = WorktreeModeCreate
				m.pathInput.Focus()
				return m, nil
			}

		case "d":
			if m.mode == WorktreeModeList && len(m.worktrees) > 0 {
				wt := m.worktrees[m.cursor]
				m.confirmPath = wt.Path
				m.confirmFocused = 1 // Default to "No"
				m.mode = WorktreeModeConfirmDelete
				m.err = ""
				return m, nil
			}

		case "j", "down":
			if m.mode == WorktreeModeList && m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}

		case "k", "up":
			if m.mode == WorktreeModeList && m.cursor > 0 {
				m.cursor--
			}

		case "tab":
			if m.mode == WorktreeModeCreate {
				m.focusedInput = 1 - m.focusedInput
				if m.focusedInput == 0 {
					m.pathInput.Focus()
					m.branchInput.Blur()
				} else {
					m.pathInput.Blur()
					m.branchInput.Focus()
				}
			}

		case "enter":
			if m.mode == WorktreeModeCreate {
				path := strings.TrimSpace(m.pathInput.Value())
				branch := strings.TrimSpace(m.branchInput.Value())

				if path == "" {
					m.err = "Path is required"
					return m, nil
				}

				if err := m.repo.AddWorktree(path, branch); err != nil {
					m.err = err.Error()
					return m, nil
				}

				m.worktrees, _ = m.repo.ListWorktrees()
				m.mode = WorktreeModeList
				m.pathInput.Reset()
				m.branchInput.Reset()
				m.err = ""
				return m, nil
			}
		}
	}

	if m.mode == WorktreeModeCreate {
		var cmd tea.Cmd
		if m.focusedInput == 0 {
			m.pathInput, cmd = m.pathInput.Update(msg)
		} else {
			m.branchInput, cmd = m.branchInput.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m *WorktreeModal) View() string {
	var title string
	var body string

	switch m.mode {
	case WorktreeModeCreate:
		title = m.styles.ModalTitle.Render(" New Worktree ")
		body = m.renderCreateView()
	case WorktreeModeConfirmDelete:
		title = m.styles.ModalTitle.Render(" Delete Worktree ")
		body = m.renderConfirmDeleteView()
	default:
		title = m.styles.ModalTitle.Render(" Worktrees ")
		body = m.renderListView()
	}

	modal := m.styles.Modal.Render(body)

	lines := strings.Split(modal, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		titleWidth := lipgloss.Width(title)
		borderStart := 2

		if len(firstLine) > borderStart+titleWidth {
			runes := []rune(firstLine)
			titleRunes := []rune(title)
			for i, r := range titleRunes {
				if borderStart+i < len(runes) {
					runes[borderStart+i] = r
				}
			}
			lines[0] = string(runes)
		}
		modal = strings.Join(lines, "\n")
	}

	return modal
}

func (m *WorktreeModal) renderListView() string {
	var lines []string

	if len(m.worktrees) == 0 {
		lines = append(lines, m.styles.Dim.Render("No worktrees"))
	} else {
		for i, wt := range m.worktrees {
			line := fmt.Sprintf("  %s", wt.Path)
			if wt.Branch != "" {
				line += fmt.Sprintf(" (%s)", wt.Branch)
			}
			if wt.Bare {
				line += " [bare]"
			}

			if i == m.cursor {
				line = m.styles.ListItemSelected.Render(line)
			} else {
				line = m.styles.ListItem.Render(line)
			}
			lines = append(lines, line)
		}
	}

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("n new • d delete • q close")

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(lines, "\n"),
		errLine,
		"",
		footer,
	)
}

func (m *WorktreeModal) renderCreateView() string {
	pathLabel := m.styles.Bold.Render("Path:")
	branchLabel := m.styles.Bold.Render("Branch (optional):")

	var errLine string
	if m.err != "" {
		errLine = "\n" + m.styles.Error.Render(m.err)
	}

	footer := m.styles.Dim.Render("Tab to switch • Enter to create • Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		pathLabel,
		m.pathInput.View(),
		"",
		branchLabel,
		m.branchInput.View(),
		errLine,
		"",
		footer,
	)
}

func (m *WorktreeModal) renderConfirmDeleteView() string {
	message := fmt.Sprintf("Delete worktree '%s'?\n\nThis cannot be undone.", m.confirmPath)
	messageView := m.styles.Base.Render(message)

	yesStyle := m.styles.Dim
	noStyle := m.styles.Dim

	if m.confirmFocused == 0 {
		yesStyle = m.styles.ListItemSelected
	} else {
		noStyle = m.styles.ListItemSelected
	}

	yes := yesStyle.Padding(0, 2).Render("Yes")
	no := noStyle.Padding(0, 2).Render("No")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yes, "  ", no)

	footer := m.styles.Dim.Render("y/n to confirm • Tab to switch • Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Center,
		messageView,
		"",
		buttons,
		"",
		footer,
	)
}
