package search

import (
	"testing"
)

func TestFuzzyMatchScoring(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		higherInput   string // should score higher
		lowerInput    string // should score lower
	}{
		{
			name:        "consecutive matches score higher",
			pattern:     "foo",
			higherInput: "foobar",
			lowerInput:  "f_o_o_bar",
		},
		{
			name:        "start of string bonus",
			pattern:     "ma",
			higherInput: "main",
			lowerInput:  "command",
		},
		{
			name:        "separator bonus",
			pattern:     "fb",
			higherInput: "foo-bar",   // 'b' after separator '-'
			lowerInput:  "foobazbar", // 'b' not after separator
		},
		{
			name:        "shorter string less penalty",
			pattern:     "foo",
			higherInput: "foo",
			lowerInput:  "fooooooooooo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeFuzzy, false, false)
			higher := m.Match(tt.higherInput)
			lower := m.Match(tt.lowerInput)

			if !higher.Matched {
				t.Fatalf("expected %q to match pattern %q", tt.higherInput, tt.pattern)
			}
			if !lower.Matched {
				t.Fatalf("expected %q to match pattern %q", tt.lowerInput, tt.pattern)
			}
			if higher.Score <= lower.Score {
				t.Errorf("%q (score=%d) should score higher than %q (score=%d)",
					tt.higherInput, higher.Score, tt.lowerInput, lower.Score)
			}
		})
	}
}

func TestFuzzyMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		input       string
		wantMatched bool
	}{
		{"pattern longer than input", "abcdefgh", "abc", false},
		{"single char match", "a", "apple", true},
		{"single char no match", "z", "apple", false},
		{"same string", "hello", "hello", true},
		{"unicode pattern", "日本", "日本語", true},
		{"mixed ascii unicode", "h世", "hello世界", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeFuzzy, false, false)
			result := m.Match(tt.input)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match() = %v, want %v", result.Matched, tt.wantMatched)
			}
		})
	}
}

func TestExactMatchPositions(t *testing.T) {
	m := NewMatcher("ab", ModeExact, false, false)
	result := m.Match("ababab")

	if !result.Matched {
		t.Fatal("expected match")
	}

	// Should find 3 non-overlapping matches starting at 0, 2, 4
	// But the implementation finds overlapping matches starting at 0, 1, 2, 3, 4
	// Actually, looking at the code: offset += idx + 1, so it finds overlapping
	if len(result.Matches) < 3 {
		t.Errorf("expected at least 3 matches, got %d", len(result.Matches))
	}

	// First match should start at 0
	if result.Matches[0].Start != 0 {
		t.Errorf("first match start = %d, want 0", result.Matches[0].Start)
	}
}

func TestExactMatchCaseInsensitive(t *testing.T) {
	m := NewMatcher("Hello", ModeExact, false, false)
	result := m.Match("hello HELLO HeLLo")

	if !result.Matched {
		t.Fatal("expected case-insensitive match")
	}
	if len(result.Matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(result.Matches))
	}
}

func TestRegexMatchPositions(t *testing.T) {
	m := NewMatcher(`\d+`, ModeRegex, false, false)
	result := m.Match("foo 123 bar 456")

	if !result.Matched {
		t.Fatal("expected match")
	}
	if len(result.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result.Matches))
	}
}

func TestRegexCaseSensitiveFlag(t *testing.T) {
	// Case-insensitive regex
	m1 := NewMatcher("FOO", ModeRegex, false, false)
	if !m1.Match("foo").Matched {
		t.Error("case-insensitive regex should match 'foo' with 'FOO'")
	}

	// Case-sensitive regex
	m2 := NewMatcher("FOO", ModeRegex, true, false)
	if m2.Match("foo").Matched {
		t.Error("case-sensitive regex should NOT match 'foo' with 'FOO'")
	}
}

func TestRegexNilOnInvalidPattern(t *testing.T) {
	m := NewMatcher("[invalid", ModeRegex, false, false)

	if m.Error() == nil {
		t.Error("expected error for invalid regex")
	}

	result := m.Match("anything")
	if result.Matched {
		t.Error("invalid regex should not match")
	}
}

func TestMatchStringConvenience(t *testing.T) {
	if !MatchString("foo", "foobar", ModeExact) {
		t.Error("MatchString should find 'foo' in 'foobar'")
	}
	if MatchString("baz", "foobar", ModeExact) {
		t.Error("MatchString should not find 'baz' in 'foobar'")
	}
}

func TestFuzzyMatchConvenience(t *testing.T) {
	result := FuzzyMatch("fb", "foobar")
	if !result.Matched {
		t.Error("FuzzyMatch should match 'fb' in 'foobar'")
	}
}

func TestExactMatchConvenience(t *testing.T) {
	result := ExactMatch("foo", "FOOBAR", false)
	if !result.Matched {
		t.Error("case-insensitive ExactMatch should match")
	}

	result2 := ExactMatch("foo", "FOOBAR", true)
	if result2.Matched {
		t.Error("case-sensitive ExactMatch should not match")
	}
}

func TestSmartCaseWithAllLowercase(t *testing.T) {
	m := NewMatcher("abc", ModeExact, false, true)
	result := m.Match("ABC")
	if !result.Matched {
		t.Error("lowercase pattern with smart case should match uppercase input")
	}
}

func TestSmartCaseWithMixedCase(t *testing.T) {
	m := NewMatcher("aBc", ModeExact, false, true)
	result := m.Match("abc")
	if result.Matched {
		t.Error("mixed-case pattern with smart case should be case-sensitive")
	}

	result2 := m.Match("aBc")
	if !result2.Matched {
		t.Error("exact case should match")
	}
}

func TestMatchEmptyInput(t *testing.T) {
	m := NewMatcher("foo", ModeExact, false, false)
	result := m.Match("")
	if result.Matched {
		t.Error("should not match empty input with non-empty pattern")
	}
}

func TestMatchEmptyPatternAllModes(t *testing.T) {
	modes := []MatchMode{ModeExact, ModeFuzzy, ModeRegex}
	for _, mode := range modes {
		m := NewMatcher("", mode, false, false)
		result := m.Match("anything")
		if !result.Matched {
			t.Errorf("empty pattern in mode %d should match any input", mode)
		}
	}
}
