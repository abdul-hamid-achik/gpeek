package panels

import tea "charm.land/bubbletea/v2"

type Panel interface {
	Update(msg tea.Msg) tea.Cmd
	View() string
	Focus()
	Blur()
	SetSize(width, height int)
	IsFocused() bool
}

type BasePanel struct {
	width   int
	height  int
	focused bool
}

func (b *BasePanel) SetSize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	b.width = width
	b.height = height
}

func (b *BasePanel) Focus() {
	b.focused = true
}

func (b *BasePanel) Blur() {
	b.focused = false
}

func (b *BasePanel) IsFocused() bool {
	return b.focused
}

func (b *BasePanel) Width() int {
	return b.width
}

func (b *BasePanel) Height() int {
	return b.height
}
