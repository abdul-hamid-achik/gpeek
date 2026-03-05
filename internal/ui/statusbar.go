package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	styles *Styles
	width  int

	RepoName    string
	Branch      string
	Ahead       int
	Behind      int
	Staged      int
	Unstaged    int
	Untracked   int
	Message     string
	IsError     bool
	HelpHints   []string
}

func NewStatusBar(styles *Styles) *StatusBar {
	return &StatusBar{
		styles: styles,
	}
}

func (s *StatusBar) SetSize(width int) {
	s.width = width
}

func (s *StatusBar) SetRepo(name, branch string) {
	s.RepoName = name
	s.Branch = branch
}

func (s *StatusBar) SetCounts(ahead, behind, staged, unstaged, untracked int) {
	s.Ahead = ahead
	s.Behind = behind
	s.Staged = staged
	s.Unstaged = unstaged
	s.Untracked = untracked
}

func (s *StatusBar) SetMessage(msg string, isError bool) {
	s.Message = msg
	s.IsError = isError
}

func (s *StatusBar) ClearMessage() {
	s.Message = ""
	s.IsError = false
}

func (s *StatusBar) SetHelpHints(hints []string) {
	s.HelpHints = hints
}

func (s *StatusBar) View() string {
	left := s.renderLeft()
	right := s.renderRight()

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := s.width - leftWidth - rightWidth

	if padding < 0 {
		padding = 0
	}

	spacer := lipgloss.NewStyle().Width(padding).Render("")

	return s.styles.StatusBar.Width(s.width).Render(left + spacer + right)
}

func (s *StatusBar) renderLeft() string {
	parts := []string{}

	if s.RepoName != "" {
		parts = append(parts, s.styles.StatusBarKey.Render(s.RepoName))
	}

	if s.Branch != "" {
		parts = append(parts, s.styles.StatusBarValue.Render(" on "))
		parts = append(parts, s.styles.StatusBarKey.Render(s.Branch))
	}

	if s.Ahead > 0 || s.Behind > 0 {
		syncStatus := ""
		if s.Ahead > 0 {
			syncStatus += fmt.Sprintf("↑%d", s.Ahead)
		}
		if s.Behind > 0 {
			if syncStatus != "" {
				syncStatus += " "
			}
			syncStatus += fmt.Sprintf("↓%d", s.Behind)
		}
		parts = append(parts, s.styles.StatusBarValue.Render(" ["))
		parts = append(parts, s.styles.Info.Render(syncStatus))
		parts = append(parts, s.styles.StatusBarValue.Render("]"))
	}

	if s.Staged > 0 || s.Unstaged > 0 || s.Untracked > 0 {
		parts = append(parts, s.styles.StatusBarValue.Render(" │ "))

		if s.Staged > 0 {
			parts = append(parts, s.styles.Added.Render(fmt.Sprintf("+%d", s.Staged)))
		}
		if s.Unstaged > 0 {
			if s.Staged > 0 {
				parts = append(parts, s.styles.StatusBarValue.Render(" "))
			}
			parts = append(parts, s.styles.Modified.Render(fmt.Sprintf("~%d", s.Unstaged)))
		}
		if s.Untracked > 0 {
			if s.Staged > 0 || s.Unstaged > 0 {
				parts = append(parts, s.styles.StatusBarValue.Render(" "))
			}
			parts = append(parts, s.styles.Untracked.Render(fmt.Sprintf("?%d", s.Untracked)))
		}
	}

	if s.Message != "" {
		parts = append(parts, s.styles.StatusBarValue.Render(" │ "))
		if s.IsError {
			parts = append(parts, s.styles.StatusBarError.Render(s.Message))
		} else {
			parts = append(parts, s.styles.StatusBarValue.Render(s.Message))
		}
	}

	result := ""
	for _, p := range parts {
		result += p
	}
	return result
}

func (s *StatusBar) renderRight() string {
	hints := s.styles.HelpKey.Render("?") + s.styles.HelpDesc.Render(" help  ") +
		s.styles.HelpKey.Render("q") + s.styles.HelpDesc.Render(" quit")
	return hints
}
