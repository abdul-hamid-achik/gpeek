package search

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Highlighter applies visual highlighting to matched text
type Highlighter struct {
	matchStyle   lipgloss.Style
	normalStyle  lipgloss.Style
}

// NewHighlighter creates a highlighter with the given styles
func NewHighlighter(matchStyle, normalStyle lipgloss.Style) *Highlighter {
	return &Highlighter{
		matchStyle:  matchStyle,
		normalStyle: normalStyle,
	}
}

// Highlight applies highlighting to a string based on match positions
func (h *Highlighter) Highlight(text string, matches []Match) string {
	if len(matches) == 0 {
		return h.normalStyle.Render(text)
	}

	// Sort and merge overlapping matches
	matches = mergeMatches(matches)

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		// Bounds check
		if match.Start < 0 || match.End > len(text) || match.Start >= match.End {
			continue
		}

		// Render text before this match
		if match.Start > lastEnd {
			result.WriteString(h.normalStyle.Render(text[lastEnd:match.Start]))
		}

		// Render the match
		result.WriteString(h.matchStyle.Render(text[match.Start:match.End]))
		lastEnd = match.End
	}

	// Render remaining text
	if lastEnd < len(text) {
		result.WriteString(h.normalStyle.Render(text[lastEnd:]))
	}

	return result.String()
}

// HighlightLine highlights a line and ensures consistent width
func (h *Highlighter) HighlightLine(text string, matches []Match, width int) string {
	highlighted := h.Highlight(text, matches)

	// Pad to width if needed (visual width)
	visWidth := lipgloss.Width(highlighted)
	if visWidth < width {
		highlighted += strings.Repeat(" ", width-visWidth)
	}

	return highlighted
}

// mergeMatches sorts matches and merges overlapping ones
func mergeMatches(matches []Match) []Match {
	if len(matches) <= 1 {
		return matches
	}

	// Simple bubble sort for small slices (typically <10 matches)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Start < matches[i].Start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Merge overlapping
	result := []Match{matches[0]}
	for i := 1; i < len(matches); i++ {
		last := &result[len(result)-1]
		if matches[i].Start <= last.End {
			// Overlapping or adjacent, merge
			if matches[i].End > last.End {
				last.End = matches[i].End
			}
		} else {
			result = append(result, matches[i])
		}
	}

	return result
}

// DefaultMatchStyle returns a common match highlight style
func DefaultMatchStyle(theme interface{ Primary() string }) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5E81AC"))
}

// HighlightMatches is a convenience function that creates a highlighter and applies it
func HighlightMatches(text string, matches []Match, matchColor, normalColor string) string {
	matchStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(normalColor)).
		Background(lipgloss.Color(matchColor))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(normalColor))

	h := NewHighlighter(matchStyle, normalStyle)
	return h.Highlight(text, matches)
}

// HighlightSimple highlights all occurrences of a pattern in text (exact match)
func HighlightSimple(text, pattern string, matchStyle, normalStyle lipgloss.Style) string {
	if pattern == "" {
		return normalStyle.Render(text)
	}

	m := NewMatcher(pattern, ModeExact, false, false)
	result := m.Match(text)

	h := NewHighlighter(matchStyle, normalStyle)
	return h.Highlight(text, result.Matches)
}
