package search

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestMergeMatches(t *testing.T) {
	tests := []struct {
		name     string
		input    []Match
		expected []Match
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty",
			input:    []Match{},
			expected: []Match{},
		},
		{
			name:     "single match",
			input:    []Match{{Start: 0, End: 3}},
			expected: []Match{{Start: 0, End: 3}},
		},
		{
			name:     "non-overlapping",
			input:    []Match{{Start: 0, End: 3}, {Start: 5, End: 8}},
			expected: []Match{{Start: 0, End: 3}, {Start: 5, End: 8}},
		},
		{
			name:     "overlapping",
			input:    []Match{{Start: 0, End: 5}, {Start: 3, End: 8}},
			expected: []Match{{Start: 0, End: 8}},
		},
		{
			name:     "adjacent",
			input:    []Match{{Start: 0, End: 3}, {Start: 3, End: 6}},
			expected: []Match{{Start: 0, End: 6}},
		},
		{
			name:     "unsorted input",
			input:    []Match{{Start: 5, End: 8}, {Start: 0, End: 3}},
			expected: []Match{{Start: 0, End: 3}, {Start: 5, End: 8}},
		},
		{
			name:     "contained within",
			input:    []Match{{Start: 0, End: 10}, {Start: 2, End: 5}},
			expected: []Match{{Start: 0, End: 10}},
		},
		{
			name: "three overlapping",
			input: []Match{
				{Start: 0, End: 4},
				{Start: 3, End: 7},
				{Start: 6, End: 10},
			},
			expected: []Match{{Start: 0, End: 10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeMatches(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("len = %d, want %d", len(result), len(tt.expected))
			}
			for i, m := range result {
				if m.Start != tt.expected[i].Start || m.End != tt.expected[i].End {
					t.Errorf("match[%d] = {%d,%d}, want {%d,%d}",
						i, m.Start, m.End, tt.expected[i].Start, tt.expected[i].End)
				}
			}
		})
	}
}

func TestHighlighterNoMatches(t *testing.T) {
	h := NewHighlighter(
		testStyle(),
		testStyle(),
	)

	result := h.Highlight("hello world", nil)
	// With no matches, the entire text should be rendered with normal style
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestHighlighterInvalidBounds(t *testing.T) {
	h := NewHighlighter(testStyle(), testStyle())

	// Matches with invalid bounds should be skipped
	matches := []Match{
		{Start: -1, End: 3},   // negative start
		{Start: 5, End: 3},    // start >= end
		{Start: 0, End: 1000}, // end > len(text)
	}

	// Should not panic
	result := h.Highlight("hello", matches)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestHighlighterWithMatches(t *testing.T) {
	h := NewHighlighter(testStyle(), testStyle())

	matches := []Match{{Start: 0, End: 5}}
	result := h.Highlight("hello world", matches)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestHighlightSimpleEmpty(t *testing.T) {
	result := HighlightSimple("hello", "", testStyle(), testStyle())
	if result == "" {
		t.Error("expected non-empty result for empty pattern")
	}
}

func TestHighlightMatchesConvenience(t *testing.T) {
	matches := []Match{{Start: 0, End: 3}}
	result := HighlightMatches("foobar", matches, "#FF0000", "#FFFFFF")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// testStyle returns a zero-value lipgloss.Style for testing
func testStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}
