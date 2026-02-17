package search

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		opts        QueryOptions
		wantText    string
		wantMode    MatchMode
		wantNegated bool
	}{
		{
			name:     "empty input",
			input:    "",
			opts:     DefaultQueryOptions(),
			wantText: "",
			wantMode: ModeFuzzy,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			opts:     DefaultQueryOptions(),
			wantText: "",
			wantMode: ModeFuzzy,
		},
		{
			name:     "plain text",
			input:    "hello",
			opts:     DefaultQueryOptions(),
			wantText: "hello",
			wantMode: ModeFuzzy,
		},
		{
			name:        "negation prefix",
			input:       "!error",
			opts:        DefaultQueryOptions(),
			wantText:    "error",
			wantMode:    ModeFuzzy,
			wantNegated: true,
		},
		{
			name:        "negation with space",
			input:       "! error",
			opts:        DefaultQueryOptions(),
			wantText:    "error",
			wantMode:    ModeFuzzy,
			wantNegated: true,
		},
		{
			name:     "regex mode /pattern/",
			input:    "/foo.*/",
			opts:     DefaultQueryOptions(),
			wantText: "foo.*",
			wantMode: ModeRegex,
		},
		{
			name:     "exact mode quoted",
			input:    `"exact match"`,
			opts:     DefaultQueryOptions(),
			wantText: "exact match",
			wantMode: ModeExact,
		},
		{
			name:        "negated regex",
			input:       "!/err.*/",
			opts:        DefaultQueryOptions(),
			wantText:    "err.*",
			wantMode:    ModeRegex,
			wantNegated: true,
		},
		{
			name:        "negated exact",
			input:       `!"foo bar"`,
			opts:        DefaultQueryOptions(),
			wantText:    "foo bar",
			wantMode:    ModeExact,
			wantNegated: true,
		},
		{
			name:     "single slash not regex",
			input:    "/",
			opts:     DefaultQueryOptions(),
			wantText: "/",
			wantMode: ModeFuzzy,
		},
		{
			name:     "single quote not exact",
			input:    `"`,
			opts:     DefaultQueryOptions(),
			wantText: `"`,
			wantMode: ModeFuzzy,
		},
		{
			name:     "regex with slashes inside",
			input:    "/path/to/",
			opts:     DefaultQueryOptions(),
			wantText: "path/to",
			wantMode: ModeRegex,
		},
		{
			name:     "respects default mode exact",
			input:    "hello",
			opts:     QueryOptions{DefaultMode: ModeExact},
			wantText: "hello",
			wantMode: ModeExact,
		},
		{
			name:     "leading trailing spaces trimmed",
			input:    "  hello  ",
			opts:     DefaultQueryOptions(),
			wantText: "hello",
			wantMode: ModeFuzzy,
		},
		{
			name:        "just exclamation mark",
			input:       "!",
			opts:        DefaultQueryOptions(),
			wantText:    "",
			wantMode:    ModeFuzzy,
			wantNegated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := ParseQuery(tt.input, tt.opts)
			if q.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", q.Text, tt.wantText)
			}
			if q.Mode != tt.wantMode {
				t.Errorf("Mode = %d, want %d", q.Mode, tt.wantMode)
			}
			if q.Negated != tt.wantNegated {
				t.Errorf("Negated = %v, want %v", q.Negated, tt.wantNegated)
			}
		})
	}
}

func TestQueryCreateMatcher(t *testing.T) {
	q := ParseQuery("hello", DefaultQueryOptions())
	m := q.CreateMatcher()

	if m == nil {
		t.Fatal("CreateMatcher returned nil")
	}

	result := m.Match("hello world")
	if !result.Matched {
		t.Error("expected match for 'hello' in 'hello world'")
	}
}

func TestQueryRegexError(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		opts      QueryOptions
		wantError bool
	}{
		{"valid regex", "/foo.*/", DefaultQueryOptions(), false},
		{"invalid regex", "/[/", QueryOptions{DefaultMode: ModeRegex}, true},
		{"fuzzy mode no error", "hello", DefaultQueryOptions(), false},
		{"empty regex", "//", DefaultQueryOptions(), false}, // empty pattern -> no error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := ParseQuery(tt.input, tt.opts)
			err := q.RegexError()
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultQueryOptions(t *testing.T) {
	opts := DefaultQueryOptions()
	if opts.DefaultMode != ModeFuzzy {
		t.Errorf("DefaultMode = %d, want ModeFuzzy", opts.DefaultMode)
	}
	if opts.DefaultCaseSens != false {
		t.Error("DefaultCaseSens should be false")
	}
	if opts.DefaultSmartCase != true {
		t.Error("DefaultSmartCase should be true")
	}
}

func TestFilterEmpty(t *testing.T) {
	// Empty query returns all items
	items := []string{"a", "b", "c"}
	query := ParseQuery("", DefaultQueryOptions())
	result := Filter(items, query, func(s string) string { return s })

	if len(result) != 3 {
		t.Errorf("empty query should return all items, got %d", len(result))
	}
}

func TestFilterNoMatches(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	query := ParseQuery("xyz", DefaultQueryOptions())
	result := Filter(items, query, func(s string) string { return s })

	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestFilterWithNegation(t *testing.T) {
	items := []string{"error: fail", "info: ok", "error: crash", "debug: trace"}
	query := ParseQuery("!error", DefaultQueryOptions())
	result := Filter(items, query, func(s string) string { return s })

	if len(result) != 2 {
		t.Errorf("expected 2 results (excluding 'error'), got %d", len(result))
	}
	for _, r := range result {
		if contains(r, "error") {
			t.Errorf("negation filter should exclude 'error', but got %q", r)
		}
	}
}

func TestFilterWithScore(t *testing.T) {
	items := []string{"foobar", "foo", "bar", "fo"}
	query := ParseQuery("foo", DefaultQueryOptions())
	results := FilterWithScore(items, query, func(s string) string { return s })

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// "foo" should have a higher score than "foobar" (shorter match)
	var fooScore, foobarScore int
	for _, r := range results {
		if r.Item == "foo" {
			fooScore = r.Score
		}
		if r.Item == "foobar" {
			foobarScore = r.Score
		}
	}
	if fooScore <= foobarScore {
		t.Errorf("'foo' (score=%d) should score higher than 'foobar' (score=%d)", fooScore, foobarScore)
	}
}

func TestFilterWithScoreEmpty(t *testing.T) {
	items := []string{"a", "b", "c"}
	query := ParseQuery("", DefaultQueryOptions())
	results := FilterWithScore(items, query, func(s string) string { return s })

	if len(results) != 3 {
		t.Errorf("empty query should return all items with score 0, got %d", len(results))
	}
	for _, r := range results {
		if r.Score != 0 {
			t.Errorf("score for empty query should be 0, got %d", r.Score)
		}
	}
}

func TestFilterWithScoreNegation(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	query := ParseQuery("!alpha", DefaultQueryOptions())
	results := FilterWithScore(items, query, func(s string) string { return s })

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Item == "alpha" {
			t.Error("negation should exclude 'alpha'")
		}
	}
}

func TestMultiFieldMatcher(t *testing.T) {
	query := ParseQuery("main", DefaultQueryOptions())
	mfm := NewMultiFieldMatcher(query)

	// Match in first field
	result := mfm.MatchAny("main.go", "helper.go")
	if !result.Matched {
		t.Error("expected match in first field")
	}

	// Match in second field
	result2 := mfm.MatchAny("test.go", "main_test.go")
	if !result2.Matched {
		t.Error("expected match in second field")
	}

	// No match
	result3 := mfm.MatchAny("test.go", "helper.go")
	if result3.Matched {
		t.Error("expected no match")
	}
}

func TestMultiFieldMatcherNegation(t *testing.T) {
	query := ParseQuery("!error", DefaultQueryOptions())
	mfm := NewMultiFieldMatcher(query)

	// Contains "error" -> negation means no match
	result := mfm.MatchAny("error.go", "something")
	if result.Matched {
		t.Error("negated query should not match field containing 'error'")
	}

	// Does not contain "error" -> negation means match
	result2 := mfm.MatchAny("success.go", "ok")
	if !result2.Matched {
		t.Error("negated query should match fields not containing 'error'")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
