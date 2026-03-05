package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/harmonica"
)

// TransitionType defines the type of transition animation
type TransitionType int

const (
	TransitionNone TransitionType = iota
	TransitionFade
	TransitionSlideUp
	TransitionSlideDown
	TransitionSlideLeft
	TransitionSlideRight
	TransitionZoom
	TransitionPop
)

// TransitionModel manages transition animations
type TransitionModel struct {
	spring      harmonica.Spring
	position    float64
	velocity    float64
	target      float64
	active      bool
	transitionType TransitionType
	width       int
	height      int
}

// NewTransition creates a new transition model
func NewTransition(fps int) TransitionModel {
	return TransitionModel{
		spring:   harmonica.NewSpring(harmonica.FPS(fps), 6.0, 0.5),
		position: 0,
		velocity: 0,
		target:   0,
		active:   false,
	}
}

// Start begins a transition animation
func (m *TransitionModel) Start(transitionType TransitionType, width, height int) {
	m.transitionType = transitionType
	m.width = width
	m.height = height
	m.target = 1.0
	m.active = true
}

// Stop ends the transition animation
func (m *TransitionModel) Stop() {
	m.target = 0.0
	m.active = true
}

// IsComplete returns true if the transition has finished
func (m TransitionModel) IsComplete() bool {
	return !m.active && m.position < 0.01
}

// TransitionTickMsg is sent to update the animation
type TransitionTickMsg struct{}

// Update handles transition animation updates
func (m TransitionModel) Update(msg tea.Msg) (TransitionModel, tea.Cmd) {
	switch msg.(type) {
	case TransitionTickMsg:
		if !m.active {
			return m, nil
		}

		m.position, m.velocity = m.spring.Update(m.position, m.velocity, m.target)

		// Check if we've reached the target
		if m.position >= 0.99 && m.target > 0.5 {
			m.position = 1.0
			m.velocity = 0
			m.active = false
		} else if m.position <= 0.01 && m.target < 0.5 {
			m.position = 0.0
			m.velocity = 0
			m.active = false
		}

		return m, nil
	}
	return m, nil
}

// GetOffset returns the current animation offset (0.0 to 1.0)
func (m TransitionModel) GetOffset() float64 {
	return m.position
}

// ApplyTransition applies the transition effect to content
func (m TransitionModel) ApplyTransition(content string, width, height int) string {
	if m.transitionType == TransitionNone || m.position < 0.01 {
		return content
	}

	// For now, return content unchanged - full transition rendering
	// would require more complex ANSI manipulation
	return content
}

// PanelZoomModel handles panel zoom transitions
type PanelZoomModel struct {
	spring    harmonica.Spring
	scale     float64
	velocity  float64
	target    float64
	active    bool
	fromWidth int
	fromHeight int
	toWidth   int
	toHeight  int
}

// NewPanelZoom creates a new panel zoom model
func NewPanelZoom(fps int) PanelZoomModel {
	return PanelZoomModel{
		spring: harmonica.NewSpring(harmonica.FPS(fps), 8.0, 0.3),
		scale:  1.0,
	}
}

// ZoomIn starts a zoom-in animation
func (m *PanelZoomModel) ZoomIn(fromWidth, fromHeight, toWidth, toHeight int) {
	m.fromWidth = fromWidth
	m.fromHeight = fromHeight
	m.toWidth = toWidth
	m.toHeight = toHeight
	m.target = 1.0
	m.active = true
}

// ZoomOut starts a zoom-out animation
func (m *PanelZoomModel) ZoomOut() {
	m.target = 0.0
	m.active = true
}

// GetCurrentSize returns the interpolated size based on animation progress
func (m PanelZoomModel) GetCurrentSize() (int, int) {
	progress := m.scale
	return interpolate(m.fromWidth, m.toWidth, progress),
		interpolate(m.fromHeight, m.toHeight, progress)
}

func interpolate(from, to int, progress float64) int {
	return from + int(float64(to-from)*progress)
}

// Update handles zoom animation updates
func (m PanelZoomModel) Update(msg tea.Msg) (PanelZoomModel, tea.Cmd) {
	switch msg.(type) {
	case TransitionTickMsg:
		if !m.active {
			return m, nil
		}

		m.scale, m.velocity = m.spring.Update(m.scale, m.velocity, m.target)

		// Check completion
		if m.scale >= 0.99 && m.target > 0.5 {
			m.scale = 1.0
			m.velocity = 0
			m.active = false
		} else if m.scale <= 0.01 && m.target < 0.5 {
			m.scale = 0.0
			m.velocity = 0
			m.active = false
		}

		return m, nil
	}
	return m, nil
}

// FadeModel handles fade in/out transitions
type FadeModel struct {
	spring   harmonica.Spring
	opacity  float64
	velocity float64
	target   float64
	active   bool
}

// NewFade creates a new fade model
func NewFade(fps int) FadeModel {
	return FadeModel{
		spring:  harmonica.NewSpring(harmonica.FPS(fps), 5.0, 0.4),
		opacity: 0.0,
	}
}

// FadeIn starts a fade-in animation
func (m *FadeModel) FadeIn() {
	m.target = 1.0
	m.active = true
}

// FadeOut starts a fade-out animation
func (m *FadeModel) FadeOut() {
	m.target = 0.0
	m.active = true
}

// GetOpacity returns the current opacity (0.0 to 1.0)
func (m FadeModel) GetOpacity() float64 {
	return m.opacity
}

// Update handles fade animation updates
func (m FadeModel) Update(msg tea.Msg) (FadeModel, tea.Cmd) {
	switch msg.(type) {
	case TransitionTickMsg:
		if !m.active {
			return m, nil
		}

		m.opacity, m.velocity = m.spring.Update(m.opacity, m.velocity, m.target)

		// Check completion
		if m.opacity >= 0.99 && m.target > 0.5 {
			m.opacity = 1.0
			m.velocity = 0
			m.active = false
		} else if m.opacity <= 0.01 && m.target < 0.5 {
			m.opacity = 0.0
			m.velocity = 0
			m.active = false
		}

		return m, nil
	}
	return m, nil
}