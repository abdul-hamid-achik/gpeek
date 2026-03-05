package modals

import (
	"fmt"
	"strings"

	"github.com/abdul-hamid-achik/gpeek/internal/git"
	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// GitConfigModal displays and allows editing of git configuration
type GitConfigModal struct {
	BaseModal
	styles *ui.Styles
	repo   *git.Repository

	sections         []git.ConfigSection
	expandedSections map[int]bool
	cursor           int

	// Edit mode
	editing   bool
	editInput textinput.Model
	editKey   string
	editError string

	// Error state
	loadError string
}

// NewGitConfigModal creates a new git config modal
func NewGitConfigModal(styles *ui.Styles, repo *git.Repository, width, height int) *GitConfigModal {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.SetWidth(width - 20)

	m := &GitConfigModal{
		styles:           styles,
		repo:             repo,
		expandedSections: make(map[int]bool),
		editInput:        ti,
	}
	m.width = width
	m.height = height

	// Load config
	sections, err := repo.GetConfig()
	if err != nil {
		m.loadError = err.Error()
	} else {
		m.sections = sections
		// Expand all sections by default
		for i := range sections {
			m.expandedSections[i] = true
		}
	}

	return m
}

// flattenedIndex returns the total items considering expanded sections
func (m *GitConfigModal) flattenedItems() []struct {
	isSection bool
	sectionIdx int
	entryIdx   int
	entry      *git.ConfigEntry
} {
	var items []struct {
		isSection bool
		sectionIdx int
		entryIdx   int
		entry      *git.ConfigEntry
	}

	for si, section := range m.sections {
		items = append(items, struct {
			isSection bool
			sectionIdx int
			entryIdx   int
			entry      *git.ConfigEntry
		}{isSection: true, sectionIdx: si})

		if m.expandedSections[si] {
			for ei := range section.Entries {
				items = append(items, struct {
					isSection bool
					sectionIdx int
					entryIdx   int
					entry      *git.ConfigEntry
				}{isSection: false, sectionIdx: si, entryIdx: ei, entry: &m.sections[si].Entries[ei]})
			}
		}
	}

	return items
}

func (m *GitConfigModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	// Handle edit mode
	if m.editing {
		return m.updateEditMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return nil, nil

		case "j", "down":
			items := m.flattenedItems()
			if m.cursor < len(items)-1 {
				m.cursor++
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "enter":
			items := m.flattenedItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isSection {
					// Toggle section expansion
					m.expandedSections[item.sectionIdx] = !m.expandedSections[item.sectionIdx]
				} else if item.entry != nil && item.entry.Editable {
					// Enter edit mode
					m.startEditing(item.entry.Key, item.entry.Value)
				}
			}

		case " ":
			// Toggle section expansion
			items := m.flattenedItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isSection {
					m.expandedSections[item.sectionIdx] = !m.expandedSections[item.sectionIdx]
				}
			}

		case "a":
			// Toggle all sections
			allExpanded := true
			for i := range m.sections {
				if !m.expandedSections[i] {
					allExpanded = false
					break
				}
			}
			for i := range m.sections {
				m.expandedSections[i] = !allExpanded
			}

		case "g":
			m.cursor = 0

		case "G":
			items := m.flattenedItems()
			if len(items) > 0 {
				m.cursor = len(items) - 1
			}
		}
	}

	return m, nil
}

func (m *GitConfigModal) updateEditMode(msg tea.Msg) (Modal, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.editing = false
			m.editError = ""
			m.editInput.Blur()
			return m, nil

		case "enter":
			// Save the value
			newValue := m.editInput.Value()
			if err := m.repo.SetConfigValue(m.editKey, newValue); err != nil {
				m.editError = err.Error()
				return m, nil
			}

			// Reload config
			sections, err := m.repo.GetConfig()
			if err == nil {
				m.sections = sections
			}

			m.editing = false
			m.editError = ""
			m.editInput.Blur()
			return m, nil
		}
	}

	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m *GitConfigModal) startEditing(key, currentValue string) {
	m.editing = true
	m.editKey = key
	m.editError = ""
	m.editInput.SetValue(currentValue)
	m.editInput.Focus()
	m.editInput.CursorEnd()
}

func (m *GitConfigModal) View() string {
	if m.editing {
		return m.viewEditMode()
	}

	return m.viewListMode()
}

func (m *GitConfigModal) viewListMode() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary)).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Git Config (Local)")

	if m.loadError != "" {
		content := m.styles.Error.Render("Error loading config: " + m.loadError)
		body := lipgloss.JoinVertical(lipgloss.Left, title, "", content)
		return m.styles.Modal.Width(m.width).Height(m.height).Render(body)
	}

	var lines []string
	items := m.flattenedItems()

	for i, item := range items {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		if item.isSection {
			// Section header
			indicator := "▼"
			if !m.expandedSections[item.sectionIdx] {
				indicator = "▶"
			}
			section := m.sections[item.sectionIdx]
			header := fmt.Sprintf("%s%s %s", prefix, indicator, section.Name)

			if i == m.cursor {
				lines = append(lines, m.styles.ListItemSelected.Render(header))
			} else {
				lines = append(lines, m.styles.Bold.Render(header))
			}
		} else if item.entry != nil {
			// Config entry
			entry := item.entry

			// Format key and value
			keyParts := strings.SplitN(entry.Key, ".", 2)
			displayKey := entry.Key
			if len(keyParts) == 2 {
				displayKey = keyParts[1]
			}

			// Truncate value if too long
			value := entry.Value
			maxValueLen := m.width - 30
			if len(value) > maxValueLen {
				value = value[:maxValueLen-3] + "..."
			}

			line := fmt.Sprintf("%s    %s", prefix, displayKey)
			valueStyle := m.styles.Dim
			if entry.Editable {
				valueStyle = m.styles.Base
			}

			// Pad key to align values
			keyWidth := lipgloss.Width(line)
			padding := 25 - keyWidth
			if padding < 1 {
				padding = 1
			}
			line += strings.Repeat(" ", padding) + valueStyle.Render(value)

			if !entry.Editable {
				line += m.styles.Dim.Render(" (read-only)")
			}

			if i == m.cursor {
				lines = append(lines, m.styles.ListItemSelected.Render(line))
			} else {
				lines = append(lines, line)
			}
		}
	}

	content := strings.Join(lines, "\n")

	// Scroll if needed
	visibleHeight := m.height - 8
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > visibleHeight {
		start := 0
		if m.cursor > visibleHeight-3 {
			start = m.cursor - visibleHeight + 3
		}
		end := start + visibleHeight
		if end > len(contentLines) {
			end = len(contentLines)
			start = end - visibleHeight
			if start < 0 {
				start = 0
			}
		}
		content = strings.Join(contentLines[start:end], "\n")
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Muted))

	footer := footerStyle.Render("j/k nav • enter edit • space toggle • a toggle all • q close")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		content,
		"",
		footer,
	)

	return m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)
}

func (m *GitConfigModal) viewEditMode() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Primary)).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Edit Config")

	keyLine := m.styles.Bold.Render("Key:   ") + m.styles.Base.Render(m.editKey)

	// Get current value for reference
	currentValue, _ := m.repo.GetConfigValue(m.editKey)
	currentLine := m.styles.Bold.Render("Current: ") + m.styles.Dim.Render(currentValue)

	newValueLabel := m.styles.Bold.Render("New value:")
	inputView := m.editInput.View()

	// Input box styling
	inputBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.styles.Theme.Primary)).
		Padding(0, 1).
		Width(m.width - 8)

	inputBox := inputBoxStyle.Render(inputView)

	var errorLine string
	if m.editError != "" {
		errorLine = m.styles.Error.Render("Error: " + m.editError)
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Theme.Muted))

	footer := footerStyle.Render("Enter to save • Esc to cancel")

	var bodyParts []string
	bodyParts = append(bodyParts, title, "", keyLine, "", currentLine, "", newValueLabel, inputBox)
	if errorLine != "" {
		bodyParts = append(bodyParts, "", errorLine)
	}
	bodyParts = append(bodyParts, "", footer)

	body := lipgloss.JoinVertical(lipgloss.Left, bodyParts...)

	return m.styles.Modal.
		Width(m.width).
		Height(m.height).
		Render(body)
}
