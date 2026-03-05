package components

import (
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ToastType represents the type of toast notification
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastStyle defines the visual style for toast notifications
type ToastStyle struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Message   lipgloss.Style
	Icon      lipgloss.Style
}

// ToastTheme contains colors for toast notifications
type ToastTheme struct {
	Success, Warning, Error, Info, Background, Foreground string
}

// DefaultToastStyles returns default toast styles
func DefaultToastStyles(theme ToastTheme) map[ToastType]ToastStyle {
	return map[ToastType]ToastStyle{
		ToastSuccess: {
			Container: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Foreground)).
				Background(lipgloss.Color(theme.Success)).
				Padding(0, 1).
				Bold(true),
			Title:   lipgloss.NewStyle().Bold(true),
			Message: lipgloss.NewStyle(),
			Icon:    lipgloss.NewStyle().Bold(true),
		},
		ToastError: {
			Container: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color(theme.Error)).
				Padding(0, 1).
				Bold(true),
			Title:   lipgloss.NewStyle().Bold(true),
			Message: lipgloss.NewStyle(),
			Icon:    lipgloss.NewStyle().Bold(true),
		},
		ToastWarning: {
			Container: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color(theme.Warning)).
				Padding(0, 1),
			Title:   lipgloss.NewStyle().Bold(true),
			Message: lipgloss.NewStyle(),
			Icon:    lipgloss.NewStyle().Bold(true),
		},
		ToastInfo: {
			Container: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Foreground)).
				Background(lipgloss.Color(theme.Info)).
				Padding(0, 1),
			Title:   lipgloss.NewStyle().Bold(true),
			Message: lipgloss.NewStyle(),
			Icon:    lipgloss.NewStyle().Bold(true),
		},
	}
}

var toastIDCounter int64

func nextToastID() int {
	return int(atomic.AddInt64(&toastIDCounter, 1))
}

// Toast represents a single toast notification
type Toast struct {
	ID        int
	Type      ToastType
	Title     string
	Message   string
	Duration  time.Duration
	CreatedAt time.Time
	style     ToastStyle
}

// ToastModel manages multiple toast notifications
type ToastModel struct {
	toasts   []Toast
	styles   map[ToastType]ToastStyle
	position ToastPosition
	maxWidth int
}

// ToastPosition defines where toasts appear on screen
type ToastPosition int

const (
	ToastTopRight ToastPosition = iota
	ToastTopLeft
	ToastBottomRight
	ToastBottomLeft
	ToastTopCenter
	ToastBottomCenter
)

// NewToastModel creates a new toast manager
func NewToastModel(styles map[ToastType]ToastStyle) ToastModel {
	return ToastModel{
		toasts:   make([]Toast, 0),
		styles:   styles,
		position: ToastBottomRight,
	}
}

// AddToast creates and adds a new toast
func (m *ToastModel) AddToast(toastType ToastType, title, message string, duration time.Duration) Toast {
	id := nextToastID()
	style, ok := m.styles[toastType]
	if !ok {
		style = m.styles[ToastInfo]
	}

	toast := Toast{
		ID:        id,
		Type:      toastType,
		Title:     title,
		Message:   message,
		Duration:  duration,
		CreatedAt: time.Now(),
		style:     style,
	}

	m.toasts = append(m.toasts, toast)
	return toast
}

// Success creates a success toast
func (m *ToastModel) Success(title, message string) Toast {
	return m.AddToast(ToastSuccess, title, message, 3*time.Second)
}

// Error creates an error toast
func (m *ToastModel) Error(title, message string) Toast {
	return m.AddToast(ToastError, title, message, 5*time.Second)
}

// Warning creates a warning toast
func (m *ToastModel) Warning(title, message string) Toast {
	return m.AddToast(ToastWarning, title, message, 4*time.Second)
}

// Info creates an info toast
func (m *ToastModel) Info(title, message string) Toast {
	return m.AddToast(ToastInfo, title, message, 3*time.Second)
}

// RemoveToast removes a toast by ID
func (m *ToastModel) RemoveToast(id int) {
	for i, t := range m.toasts {
		if t.ID == id {
			m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
			break
		}
	}
}

// Clear removes all toasts
func (m *ToastModel) Clear() {
	m.toasts = m.toasts[:0]
}

// ToastTickMsg is sent to check for expired toasts
type ToastTickMsg struct{}

// Update handles toast expiration
func (m ToastModel) Update(msg tea.Msg) (ToastModel, tea.Cmd) {
	switch msg.(type) {
	case ToastTickMsg:
		now := time.Now()
		var remaining []Toast
		for _, t := range m.toasts {
			if now.Sub(t.CreatedAt) < t.Duration {
				remaining = append(remaining, t)
			}
		}
		m.toasts = remaining

		if len(m.toasts) > 0 {
			return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return ToastTickMsg{}
			})
		}
	}
	return m, nil
}

// View renders all toasts
func (m ToastModel) View() string {
	if len(m.toasts) == 0 {
		return ""
	}

	var rendered []string
	for _, t := range m.toasts {
		rendered = append(rendered, t.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// View renders a single toast
func (t Toast) View() string {
	icon := t.getIcon()
	content := icon + " "
	if t.Title != "" {
		content += t.Title
		if t.Message != "" {
			content += ": " + t.Message
		}
	} else {
		content += t.Message
	}

	return t.style.Container.Render(content)
}

func (t Toast) getIcon() string {
	switch t.Type {
	case ToastSuccess:
		return "✓"
	case ToastError:
		return "✗"
	case ToastWarning:
		return "⚠"
	case ToastInfo:
		return "ℹ"
	default:
		return "•"
	}
}

// SetPosition sets the toast position
func (m *ToastModel) SetPosition(pos ToastPosition) {
	m.position = pos
}

// SetMaxWidth sets the maximum width for toasts
func (m *ToastModel) SetMaxWidth(width int) {
	m.maxWidth = width
}

// HasToasts returns whether there are active toasts
func (m ToastModel) HasToasts() bool {
	return len(m.toasts) > 0
}

// Count returns the number of active toasts
func (m ToastModel) Count() int {
	return len(m.toasts)
}

// StartToastTicker returns a command to start the toast expiration ticker
func StartToastTicker() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}