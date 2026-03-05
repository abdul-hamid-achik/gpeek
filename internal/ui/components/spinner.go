package components

import (
	"fmt"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SpinnerStyle defines the visual style of the spinner
type SpinnerStyle struct {
	Spinner lipgloss.Style
	Text    lipgloss.Style
}

// DefaultSpinnerStyle returns a default spinner style
func DefaultSpinnerStyle() SpinnerStyle {
	return SpinnerStyle{
		Spinner: lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")),
		Text:    lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
	}
}

// SpinnerFrames defines different spinner animation patterns
var SpinnerFrames = struct {
	Dots    []string
	Line    []string
	Arrow   []string
	Pulse   []string
	Growing []string
}{
	Dots:    []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
	Line:    []string{"|", "/", "-", "\\"},
	Arrow:   []string{"→", "↘", "↓", "↙", "←", "↖", "↑", "↗"},
	Pulse:   []string{"█", "▓", "▒", "░"},
	Growing: []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
}

var lastSpinnerID int64

func nextSpinnerID() int {
	return int(atomic.AddInt64(&lastSpinnerID, 1))
}

// SpinnerModel represents a loading spinner component
type SpinnerModel struct {
	ID      int
	frame   int
	frames  []string
	style   SpinnerStyle
	message string
	running bool
}

// NewSpinner creates a new spinner with the given frames and style
func NewSpinner(frames []string, style SpinnerStyle) SpinnerModel {
	return SpinnerModel{
		ID:      nextSpinnerID(),
		frames:  frames,
		style:   style,
		running: false,
	}
}

// NewDotsSpinner creates a spinner with dot animation
func NewDotsSpinner(style SpinnerStyle) SpinnerModel {
	return NewSpinner(SpinnerFrames.Dots, style)
}

// NewLineSpinner creates a spinner with line animation
func NewLineSpinner(style SpinnerStyle) SpinnerModel {
	return NewSpinner(SpinnerFrames.Line, style)
}

// TickMsg is sent on each animation frame
type TickMsg struct {
	ID   int
	Time int64
}

// Start begins the spinner animation
func (m SpinnerModel) Start() tea.Cmd {
	m.running = true
	return m.tick()
}

// Stop stops the spinner animation
func (m *SpinnerModel) Stop() {
	m.running = false
}

// SetMessage sets the spinner's message
func (m *SpinnerModel) SetMessage(msg string) {
	m.message = msg
}

// Update handles messages for the spinner
func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.ID != m.ID || !m.running {
			return m, nil
		}
		m.frame = (m.frame + 1) % len(m.frames)
		return m, m.tick()
	}
	return m, nil
}

func (m SpinnerModel) tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{ID: m.ID, Time: t.UnixNano()}
	})
}

// View renders the spinner
func (m SpinnerModel) View() string {
	if m.frame >= len(m.frames) {
		return ""
	}
	spinner := m.style.Spinner.Render(m.frames[m.frame])
	if m.message != "" {
		return spinner + " " + m.style.Text.Render(m.message)
	}
	return spinner
}

// LoadingIndicator is a simpler, non-animated loading indicator
type LoadingIndicator struct {
	spinner  SpinnerModel
	style    SpinnerStyle
	Width    int
	Message  string
	Progress float64 // 0-1 for progress, -1 for indeterminate
}

// NewLoadingIndicator creates a new loading indicator
func NewLoadingIndicator(message string, style SpinnerStyle) LoadingIndicator {
	return LoadingIndicator{
		spinner:  NewDotsSpinner(style),
		style:    style,
		Message:  message,
		Progress: -1, // indeterminate by default
	}
}

// Update handles messages for the loading indicator
func (m LoadingIndicator) Update(msg tea.Msg) (LoadingIndicator, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the loading indicator
func (m LoadingIndicator) View() string {
	if m.Progress >= 0 {
		return m.renderProgress()
	}
	return m.spinner.View() + " " + m.style.Text.Render(m.Message)
}

func (m LoadingIndicator) renderProgress() string {
	width := m.Width
	if width <= 0 {
		width = 40
	}

	filled := int(float64(width) * m.Progress)
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	progressBar := m.style.Spinner.Render(bar)
	percent := fmt.Sprintf(" %.0f%%", m.Progress*100)

	return m.spinner.View() + " " + m.Message + "\n" + progressBar + percent
}

// SetProgress sets the progress (0-1)
func (m *LoadingIndicator) SetProgress(p float64) {
	m.Progress = p
}