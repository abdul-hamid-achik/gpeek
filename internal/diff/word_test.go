package diff

import (
	"testing"
)

func TestComputeWordDiff(t *testing.T) {
	tests := []struct {
		name        string
		oldLine     string
		newLine     string
		oldChanged  int
		newChanged  int
	}{
		{
			name:       "no changes",
			oldLine:    "hello world",
			newLine:    "hello world",
			oldChanged: 0,
			newChanged: 0,
		},
		{
			name:       "single word change",
			oldLine:    "hello world",
			newLine:    "hello universe",
			oldChanged: 1,
			newChanged: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeWordDiff(tt.oldLine, tt.newLine)

			oldChanged := 0
			for _, w := range result.OldWords {
				if w.IsChanged {
					oldChanged++
				}
			}

			newChanged := 0
			for _, w := range result.NewWords {
				if w.IsChanged {
					newChanged++
				}
			}

			if oldChanged != tt.oldChanged {
				t.Errorf("expected %d old changed words, got %d", tt.oldChanged, oldChanged)
			}
			if newChanged != tt.newChanged {
				t.Errorf("expected %d new changed words, got %d", tt.newChanged, newChanged)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 1},
		{"hello world", 3},
		{"foo(bar)", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(tt.input)
			if len(tokens) != tt.expected {
				t.Errorf("expected %d tokens, got %d: %v", tt.expected, len(tokens), tokens)
			}
		})
	}
}

func TestWordDiffRender(t *testing.T) {
	oldLine := "hello world"
	newLine := "hello universe"

	wd := ComputeWordDiff(oldLine, newLine)

	changedStyle := func(s string) string { return "[" + s + "]" }
	normalStyle := func(s string) string { return s }

	oldRendered := wd.RenderOld(changedStyle, normalStyle)
	newRendered := wd.RenderNew(changedStyle, normalStyle)

	if oldRendered == "" {
		t.Error("old rendered should not be empty")
	}
	if newRendered == "" {
		t.Error("new rendered should not be empty")
	}
}
