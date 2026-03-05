package components

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FormStyle defines styling for form components
type FormStyle struct {
	Label         lipgloss.Style
	Input         lipgloss.Style
	InputFocused  lipgloss.Style
	Error         lipgloss.Style
	Help          lipgloss.Style
	GroupTitle    lipgloss.Style
	Required      lipgloss.Style
	Checkbox      lipgloss.Style
	CheckboxChecked lipgloss.Style
}

// DefaultFormStyle returns default form styling
func DefaultFormStyle() FormStyle {
	return FormStyle{
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true),
		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#313244")).
			Padding(0, 1),
		InputFocused: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f38ba8")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")),
		GroupTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")).
			Bold(true).
			Underline(true),
		Required: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f38ba8")),
		Checkbox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")),
		CheckboxChecked: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1")),
	}
}

// FieldType defines the type of form field
type FieldType int

const (
	FieldTypeText FieldType = iota
	FieldTypePassword
	FieldTypeNumber
	FieldTypeTextarea
	FieldTypeCheckbox
	FieldTypeSelect
	FieldTypeMultiSelect
)

// FieldOption configures a form field
type FieldOption func(*FormField)

// FormField represents a single form field
type FormField struct {
	Key         string
	Label       string
	Placeholder string
	Type        FieldType
	Required    bool
	Options     []string // For select/multi-select
	Value       string
	Error       string
	HelpText    string
	input       textinput.Model
	focused     bool
	width       int
}

// NewFormField creates a new form field
func NewFormField(key, label string, fieldType FieldType, opts ...FieldOption) FormField {
	ti := textinput.New()
	ti.Placeholder = label

	field := FormField{
		Key:     key,
		Label:   label,
		Type:    fieldType,
		input:   ti,
		Options: []string{},
	}

	for _, opt := range opts {
		opt(&field)
	}

	if field.Placeholder != "" {
		ti.Placeholder = field.Placeholder
	}

	if fieldType == FieldTypePassword {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '*'
	}

	return field
}

// WithPlaceholder sets the placeholder text
func WithPlaceholder(placeholder string) FieldOption {
	return func(f *FormField) {
		f.Placeholder = placeholder
	}
}

// WithRequired marks the field as required
func WithRequired() FieldOption {
	return func(f *FormField) {
		f.Required = true
	}
}

// WithOptions sets options for select fields
func WithOptions(options ...string) FieldOption {
	return func(f *FormField) {
		f.Options = options
	}
}

// WithHelpText sets help text for the field
func WithHelpText(help string) FieldOption {
	return func(f *FormField) {
		f.HelpText = help
	}
}

// WithInitialValue sets the initial value
func WithInitialValue(value string) FieldOption {
	return func(f *FormField) {
		f.Value = value
		f.input.SetValue(value)
	}
}

// Focus focuses the field
func (f *FormField) Focus() {
	f.focused = true
	f.input.Focus()
}

// Blur removes focus from the field
func (f *FormField) Blur() {
	f.focused = false
	f.input.Blur()
}

// SetWidth sets the field width
func (f *FormField) SetWidth(width int) {
	f.width = width
	f.input.SetWidth(width - 2) // Account for border
}

// Update handles input for the field
func (f *FormField) Update(msg tea.Msg) tea.Cmd {
	if f.focused {
		var cmd tea.Cmd
		f.input, cmd = f.input.Update(msg)
		f.Value = f.input.Value()
		return cmd
	}
	return nil
}

// View renders the field
func (f *FormField) View(style FormStyle) string {
	var b strings.Builder

	// Label
	label := f.Label
	if f.Required {
		label += style.Required.Render(" *")
	}
	b.WriteString(style.Label.Render(label))
	b.WriteString("\n")

	// Input
	inputStyle := style.Input
	if f.focused {
		inputStyle = style.InputFocused
	}

	switch f.Type {
	case FieldTypeCheckbox:
		check := "[ ]"
		if f.Value == "true" || f.Value == "on" {
			check = style.CheckboxChecked.Render("[✓]")
		} else {
			check = style.Checkbox.Render("[ ]")
		}
		b.WriteString(check + " " + f.Label)
	case FieldTypeSelect:
		// Show current selection or placeholder
		display := f.Placeholder
		if f.Value != "" {
			display = f.Value
		}
		b.WriteString(inputStyle.Width(f.width).Render(display))
	default:
		b.WriteString(inputStyle.Width(f.width).Render(f.input.View()))
	}

	// Error
	if f.Error != "" {
		b.WriteString("\n")
		b.WriteString(style.Error.Render("  ⚠ " + f.Error))
	}

	// Help
	if f.HelpText != "" && f.Error == "" {
		b.WriteString("\n")
		b.WriteString(style.Help.Render("  " + f.HelpText))
	}

	return b.String()
}

// Validate validates the field
func (f *FormField) Validate() bool {
	f.Error = ""

	if f.Required && strings.TrimSpace(f.Value) == "" {
		f.Error = "This field is required"
		return false
	}

	return true
}

// FormModel represents a complete form
type FormModel struct {
	fields    []FormField
	focusIdx  int
	width     int
	style     FormStyle
	title     string
	submitCmd func(map[string]string) tea.Cmd
}

// NewForm creates a new form
func NewForm(title string, style FormStyle) FormModel {
	return FormModel{
		title:    title,
		style:    style,
		fields:   []FormField{},
		focusIdx: 0,
	}
}

// AddField adds a field to the form
func (m *FormModel) AddField(field FormField) {
	m.fields = append(m.fields, field)
}

// SetWidth sets the form width
func (m *FormModel) SetWidth(width int) {
	m.width = width
	for i := range m.fields {
		m.fields[i].SetWidth(width)
	}
}

// SetSubmit sets the submit callback
func (m *FormModel) SetSubmit(cmd func(map[string]string) tea.Cmd) {
	m.submitCmd = cmd
}

// Init initializes the form
func (m FormModel) Init() tea.Cmd {
	if len(m.fields) > 0 {
		m.fields[0].Focus()
	}
	return nil
}

// Update handles form updates
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			m.nextField()
		case "shift+tab", "up":
			m.prevField()
		case "enter":
			if m.focusIdx == len(m.fields)-1 {
				// Submit form
				return m, m.submit()
			}
			m.nextField()
		case "ctrl+enter":
			return m, m.submit()
		default:
			if m.focusIdx < len(m.fields) {
				cmd := m.fields[m.focusIdx].Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	default:
		if m.focusIdx < len(m.fields) {
			cmd := m.fields[m.focusIdx].Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *FormModel) nextField() {
	if len(m.fields) == 0 {
		return
	}
	m.fields[m.focusIdx].Blur()
	m.focusIdx = (m.focusIdx + 1) % len(m.fields)
	m.fields[m.focusIdx].Focus()
}

func (m *FormModel) prevField() {
	if len(m.fields) == 0 {
		return
	}
	m.fields[m.focusIdx].Blur()
	m.focusIdx--
	if m.focusIdx < 0 {
		m.focusIdx = len(m.fields) - 1
	}
	m.fields[m.focusIdx].Focus()
}

func (m FormModel) submit() tea.Cmd {
	if m.submitCmd == nil {
		return nil
	}

	values := make(map[string]string)
	for _, f := range m.fields {
		values[f.Key] = f.Value
	}

	return m.submitCmd(values)
}

// Validate validates all fields
func (m FormModel) Validate() bool {
	valid := true
	for i := range m.fields {
		if !m.fields[i].Validate() {
			valid = false
		}
	}
	return valid
}

// View renders the form
func (m FormModel) View() string {
	var b strings.Builder

	if m.title != "" {
		b.WriteString(m.style.GroupTitle.Render(m.title))
		b.WriteString("\n\n")
	}

	for i, f := range m.fields {
		b.WriteString(f.View(m.style))
		if i < len(m.fields)-1 {
			b.WriteString("\n\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(m.style.Help.Render("Tab/Shift+Tab: switch fields • Enter: next/submit • Ctrl+Enter: submit"))

	return b.String()
}

// GetValues returns all field values
func (m FormModel) GetValues() map[string]string {
	values := make(map[string]string)
	for _, f := range m.fields {
		values[f.Key] = f.Value
	}
	return values
}

// SetValues sets field values
func (m *FormModel) SetValues(values map[string]string) {
	for i := range m.fields {
		if v, ok := values[m.fields[i].Key]; ok {
			m.fields[i].Value = v
			m.fields[i].input.SetValue(v)
		}
	}
}

// MultiSelectField handles multi-selection
type MultiSelectField struct {
	Label    string
	Options  []string
	Selected map[int]bool
	focused  int
	width    int
	style    FormStyle
}

// NewMultiSelect creates a multi-select field
func NewMultiSelect(label string, options []string, style FormStyle) MultiSelectField {
	return MultiSelectField{
		Label:    label,
		Options:  options,
		Selected: make(map[int]bool),
		style:    style,
	}
}

// Update handles input
func (m MultiSelectField) Update(msg tea.Msg) (MultiSelectField, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.focused > 0 {
				m.focused--
			}
		case "down", "j":
			if m.focused < len(m.Options)-1 {
				m.focused++
			}
		case " ", "x":
			m.Selected[m.focused] = !m.Selected[m.focused]
		}
	}
	return m, nil
}

// View renders the multi-select
func (m MultiSelectField) View() string {
	var b strings.Builder

	b.WriteString(m.style.Label.Render(m.Label))
	b.WriteString("\n")

	for i, opt := range m.Options {
		prefix := "[ ] "
		style := m.style.Checkbox

		if m.Selected[i] {
			prefix = "[✓] "
			style = m.style.CheckboxChecked
		}

		line := prefix + opt
		if i == m.focused {
			line = "> " + line
		} else {
			line = "  " + line
		}

		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

// GetSelected returns selected options
func (m MultiSelectField) GetSelected() []string {
	var selected []string
	for i, opt := range m.Options {
		if m.Selected[i] {
			selected = append(selected, opt)
		}
	}
	return selected
}

// ConfirmDialog shows a confirmation prompt
type ConfirmDialog struct {
	Title   string
	Message string
	style   FormStyle
	focused int // 0 = confirm, 1 = cancel
}

// NewConfirmDialog creates a confirmation dialog
func NewConfirmDialog(title, message string, style FormStyle) ConfirmDialog {
	return ConfirmDialog{
		Title:   title,
		Message: message,
		style:   style,
		focused: 0,
	}
}

// Update handles input
func (m ConfirmDialog) Update(msg tea.Msg) (ConfirmDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "left", "h", "tab":
			m.focused = (m.focused + 1) % 2
		case "right", "l", "shift+tab":
			m.focused = (m.focused + 1) % 2
		}
	}
	return m, nil
}

// View renders the dialog
func (m ConfirmDialog) View() string {
	var b strings.Builder

	b.WriteString(m.style.GroupTitle.Render(m.Title))
	b.WriteString("\n\n")
	b.WriteString(m.Message)
	b.WriteString("\n\n")

	confirmStyle := m.style.Input
	cancelStyle := m.style.Input

	if m.focused == 0 {
		confirmStyle = m.style.InputFocused
	} else {
		cancelStyle = m.style.InputFocused
	}

	buttons := confirmStyle.Render(" Confirm ") + "  " + cancelStyle.Render(" Cancel ")
	b.WriteString(buttons)

	return b.String()
}

// IsConfirmed returns true if confirm is selected
func (m ConfirmDialog) IsConfirmed() bool {
	return m.focused == 0
}