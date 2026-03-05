package modals

import (
	tea "charm.land/bubbletea/v2"
)

type Modal interface {
	Update(msg tea.Msg) (Modal, tea.Cmd)
	View() string
	ShouldClose() bool
}

type BaseModal struct {
	width  int
	height int
	closed bool
}

func (m *BaseModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *BaseModal) ShouldClose() bool {
	return m.closed
}

func (m *BaseModal) Close() {
	m.closed = true
}

func (m *BaseModal) Width() int {
	return m.width
}

func (m *BaseModal) Height() int {
	return m.height
}
