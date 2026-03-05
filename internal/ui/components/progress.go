package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ProgressStyle defines the visual style for progress bars
type ProgressStyle struct {
	Container   lipgloss.Style
	Bar         lipgloss.Style
	BarFilled   lipgloss.Style
	BarEmpty    lipgloss.Style
	Label       lipgloss.Style
	Percentage  lipgloss.Style
	Description lipgloss.Style
}

// DefaultProgressStyle returns a default progress style
func DefaultProgressStyle() ProgressStyle {
	return ProgressStyle{
		Container: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")),
		Bar: lipgloss.NewStyle(),
		BarFilled: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")),
		BarEmpty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#313244")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		Percentage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Bold(true),
		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")),
	}
}

// ProgressModel represents a progress bar component
type ProgressModel struct {
	Width       int
	Progress    float64 // 0.0 to 1.0
	Label       string
	Description string
	ShowPercent bool
	Animated    bool
	frame       int
	style       ProgressStyle
}

// NewProgress creates a new progress bar
func NewProgress(width int, style ProgressStyle) ProgressModel {
	return ProgressModel{
		Width:       width,
		Progress:    0,
		ShowPercent: true,
		Animated:    true,
		style:       style,
	}
}

// SetProgress sets the current progress (0-1)
func (m *ProgressModel) SetProgress(p float64) {
	if p < 0 {
		p = 0
	} else if p > 1 {
		p = 1
	}
	m.Progress = p
}

// SetLabel sets the progress bar label
func (m *ProgressModel) SetLabel(label string) {
	m.Label = label
}

// SetDescription sets the description text
func (m *ProgressModel) SetDescription(desc string) {
	m.Description = desc
}

// ProgressTickMsg is sent for animation updates
type ProgressTickMsg struct{}

// Update handles progress bar animation
func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg.(type) {
	case ProgressTickMsg:
		if m.Animated {
			m.frame = (m.frame + 1) % 4
		}
		return m, nil
	}
	return m, nil
}

// View renders the progress bar
func (m ProgressModel) View() string {
	barWidth := m.Width
	if barWidth <= 0 {
		barWidth = 40
	}

	var parts []string

	// Add label if present
	if m.Label != "" {
		parts = append(parts, m.style.Label.Render(m.Label))
	}

	// Render progress bar
	bar := m.renderBar(barWidth)
	parts = append(parts, bar)

	// Add percentage if enabled
	if m.ShowPercent {
		percent := m.style.Percentage.Render(fmt.Sprintf(" %.0f%%", m.Progress*100))
		parts = append(parts, percent)
	}

	// Add description if present
	if m.Description != "" {
		parts = append(parts, "\n"+m.style.Description.Render(m.Description))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func (m ProgressModel) renderBar(width int) string {
	filled := int(float64(width) * m.Progress)
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += m.style.BarFilled.Render("█")
		} else {
			bar += m.style.BarEmpty.Render("░")
		}
	}

	return m.style.Bar.Render(bar)
}

// ViewCompact renders a compact progress bar (just the bar)
func (m ProgressModel) ViewCompact() string {
	return m.renderBar(m.Width)
}

// ProgressWithIndeterminate shows an indeterminate progress animation
type IndeterminateProgress struct {
	Width    int
	frame    int
	running  bool
	style    ProgressStyle
	Label    string
}

// NewIndeterminateProgress creates an indeterminate progress bar
func NewIndeterminateProgress(width int, style ProgressStyle) IndeterminateProgress {
	return IndeterminateProgress{
		Width:   width,
		running: false,
		style:   style,
	}
}

// Start begins the indeterminate animation
func (m *IndeterminateProgress) Start() {
	m.running = true
}

// Stop stops the animation
func (m *IndeterminateProgress) Stop() {
	m.running = false
}

// Update handles animation frames
func (m IndeterminateProgress) Update(msg tea.Msg) (IndeterminateProgress, tea.Cmd) {
	switch msg.(type) {
	case ProgressTickMsg:
		if m.running {
			m.frame = (m.frame + 1) % (m.Width / 2)
		}
		return m, nil
	}
	return m, nil
}

// View renders the indeterminate progress bar
func (m IndeterminateProgress) View() string {
	barWidth := m.Width
	if barWidth <= 0 {
		barWidth = 40
	}

	bar := ""
	scannerPos := m.frame * 2
	scannerWidth := 6

	for i := 0; i < barWidth; i++ {
		if i >= scannerPos && i < scannerPos+scannerWidth {
			bar += m.style.BarFilled.Render("█")
		} else {
			bar += m.style.BarEmpty.Render("░")
		}
	}

	if m.Label != "" {
		return m.style.Label.Render(m.Label) + "\n" + bar
	}
	return bar
}

// StepProgress shows progress through a series of steps
type StepProgress struct {
	Steps     []string
	Current   int
	Completed bool
	Failed    int // -1 if no failure, otherwise the step index that failed
	style     ProgressStyle
}

// NewStepProgress creates a step-based progress indicator
func NewStepProgress(steps []string, style ProgressStyle) StepProgress {
	return StepProgress{
		Steps:   steps,
		Current: 0,
		Failed:  -1,
		style:   style,
	}
}

// Next advances to the next step
func (m *StepProgress) Next() {
	if m.Current < len(m.Steps)-1 {
		m.Current++
	}
}

// SetFailed marks the current step as failed
func (m *StepProgress) SetFailed() {
	m.Failed = m.Current
}

// SetCompleted marks all steps as completed
func (m *StepProgress) SetCompleted() {
	m.Current = len(m.Steps) - 1
	m.Completed = true
}

// View renders the step progress
func (m StepProgress) View() string {
	var lines []string

	for i, step := range m.Steps {
		var icon string
		var style lipgloss.Style

		switch {
		case i < m.Current || m.Completed:
			icon = "✓"
			style = m.style.BarFilled
		case i == m.Failed:
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
		case i == m.Current:
			icon = "●"
			style = m.style.Label
		default:
			icon = "○"
			style = m.style.BarEmpty
		}

		line := style.Render(icon) + " " + m.style.Container.Render(step)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}