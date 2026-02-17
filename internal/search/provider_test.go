package search

import (
	"encoding/json"
	"testing"
)

func TestParseVecgrepOutputJSON(t *testing.T) {
	input := []vecgrepResult{
		{
			FilePath: "main.go",
			Line:     10,
			Content:  "func main() {}",
			Score:    0.95,
		},
		{
			File:      "helper.go",
			StartLine: 20,
			Content:   "func helper() {}",
			Similarity: 0.85,
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	results := parseVecgrepOutput(data)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result uses FilePath and Line
	if results[0].FilePath != "main.go" {
		t.Errorf("result 0 FilePath = %q, want 'main.go'", results[0].FilePath)
	}
	if results[0].LineNumber != 10 {
		t.Errorf("result 0 LineNumber = %d, want 10", results[0].LineNumber)
	}
	if results[0].Score != 0.95 {
		t.Errorf("result 0 Score = %f, want 0.95", results[0].Score)
	}

	// Second result uses File fallback and StartLine, Similarity
	if results[1].FilePath != "helper.go" {
		t.Errorf("result 1 FilePath = %q, want 'helper.go'", results[1].FilePath)
	}
	if results[1].LineNumber != 20 {
		t.Errorf("result 1 LineNumber = %d, want 20", results[1].LineNumber)
	}
	if results[1].Score != 0.85 {
		t.Errorf("result 1 Score = %f, want 0.85", results[1].Score)
	}
	if results[1].MatchType != "semantic" {
		t.Errorf("result 1 MatchType = %q, want 'semantic'", results[1].MatchType)
	}
}

func TestParseVecgrepOutputJSONL(t *testing.T) {
	// JSONL format - one JSON object per line
	input := `{"file_path":"a.go","line":1,"content":"package a","score":0.9}
{"file_path":"b.go","line":5,"content":"func b()","score":0.8}
`

	results := parseVecgrepOutput([]byte(input))

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].FilePath != "a.go" {
		t.Errorf("result 0 FilePath = %q, want 'a.go'", results[0].FilePath)
	}
	if results[1].FilePath != "b.go" {
		t.Errorf("result 1 FilePath = %q, want 'b.go'", results[1].FilePath)
	}
}

func TestParseVecgrepOutputEmpty(t *testing.T) {
	results := parseVecgrepOutput(nil)
	if results != nil {
		t.Errorf("expected nil for empty input, got %v", results)
	}

	results2 := parseVecgrepOutput([]byte{})
	if results2 != nil {
		t.Errorf("expected nil for empty bytes, got %v", results2)
	}
}

func TestParseVecgrepOutputInvalid(t *testing.T) {
	// Invalid JSON that's not valid JSONL either
	results := parseVecgrepOutput([]byte("this is not json"))
	if len(results) != 0 {
		t.Errorf("expected 0 results for invalid input, got %d", len(results))
	}
}

func TestParseGrepOutputStandard(t *testing.T) {
	input := `main.go:10:func main() {
helper.go:20:func helper() {
utils.go:30:// utility function`

	results := parseGrepOutput([]byte(input), 100)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].FilePath != "main.go" {
		t.Errorf("result 0 FilePath = %q, want 'main.go'", results[0].FilePath)
	}
	if results[0].LineNumber != 10 {
		t.Errorf("result 0 LineNumber = %d, want 10", results[0].LineNumber)
	}
	if results[0].Content != "func main() {" {
		t.Errorf("result 0 Content = %q, want 'func main() {'", results[0].Content)
	}
	if results[0].MatchType != "keyword" {
		t.Errorf("result 0 MatchType = %q, want 'keyword'", results[0].MatchType)
	}
}

func TestParseGrepOutputWithLimit(t *testing.T) {
	input := `a.go:1:line1
b.go:2:line2
c.go:3:line3
d.go:4:line4`

	results := parseGrepOutput([]byte(input), 2)

	if len(results) != 2 {
		t.Errorf("expected 2 results (limit), got %d", len(results))
	}
}

func TestParseGrepOutputEmpty(t *testing.T) {
	results := parseGrepOutput(nil, 10)
	if results != nil {
		t.Errorf("expected nil for empty input, got %v", results)
	}
}

func TestParseGrepOutputNoLineNumber(t *testing.T) {
	// Format without line number: file:content
	input := `main.go:some content here`

	results := parseGrepOutput([]byte(input), 10)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// With standard format "file:linenum:content", this should parse as file=main.go, line=0 (NaN), content=some content here
	// But "main.go:some content here" only has 2 parts when split by ":"
	// Actually let me check: SplitN(line, ":", 3) on "main.go:some content here"
	// gives ["main.go", "some content here"] (2 parts) since there are only 1 colon
	// Wait no - there's "main.go" then "some content here", that's 2 parts.
	// The code handles len(parts) == 2 as file:content
	if results[0].FilePath != "main.go" {
		t.Errorf("FilePath = %q, want 'main.go'", results[0].FilePath)
	}
	if results[0].Content != "some content here" {
		t.Errorf("Content = %q, want 'some content here'", results[0].Content)
	}
}

func TestParseGrepOutputRipgrepJSON(t *testing.T) {
	input := `{"type":"match","data":{"path":{"text":"main.go"},"line_number":5,"lines":{"text":"func main() {\n"}}}
{"type":"context","data":{"path":{"text":"main.go"},"line_number":6}}
{"type":"match","data":{"path":{"text":"test.go"},"line_number":10,"lines":{"text":"func TestFoo() {\n"}}}`

	results := parseGrepOutput([]byte(input), 100)

	// Only "match" type should be included, not "context"
	if len(results) != 2 {
		t.Fatalf("expected 2 results (matches only), got %d", len(results))
	}

	if results[0].FilePath != "main.go" {
		t.Errorf("result 0 FilePath = %q, want 'main.go'", results[0].FilePath)
	}
	if results[0].LineNumber != 5 {
		t.Errorf("result 0 LineNumber = %d, want 5", results[0].LineNumber)
	}
	// Content should have trailing newline stripped
	if results[0].Content != "func main() {" {
		t.Errorf("result 0 Content = %q, want 'func main() {'", results[0].Content)
	}
}

func TestParseGrepOutputEmptyLines(t *testing.T) {
	input := "\n\nmain.go:1:content\n\n"
	results := parseGrepOutput([]byte(input), 10)

	if len(results) != 1 {
		t.Errorf("expected 1 result (skip empty lines), got %d", len(results))
	}
}

func TestSearchErrorType(t *testing.T) {
	err := ErrProviderUnavailable
	if err.Error() != "search provider unavailable" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestFallbackProviderAvailable(t *testing.T) {
	p := NewFallbackProvider("/tmp")
	if !p.Available() {
		t.Error("FallbackProvider should always be available")
	}
	if p.Name() != "text" {
		t.Errorf("name = %q, want 'text'", p.Name())
	}
}

func TestVecgrepProviderName(t *testing.T) {
	p := NewVecgrepProvider("/tmp")
	if p.Name() != "vecgrep" {
		t.Errorf("name = %q, want 'vecgrep'", p.Name())
	}
}

func TestMultiProviderNoProviders(t *testing.T) {
	mp := NewMultiProvider()
	if mp.Available() {
		t.Error("empty multi provider should not be available")
	}
	if mp.Name() != "none" {
		t.Errorf("name = %q, want 'none'", mp.Name())
	}
}

func TestMultiProviderWithFallback(t *testing.T) {
	fb := NewFallbackProvider("/tmp")
	mp := NewMultiProvider(fb)
	if !mp.Available() {
		t.Error("multi provider with fallback should be available")
	}
	if mp.Name() != "text" {
		t.Errorf("name = %q, want 'text'", mp.Name())
	}
}

func TestParseGrepOutputContentWithColons(t *testing.T) {
	// Content that contains colons should be preserved
	input := `config.go:15:host: "localhost:8080"`

	results := parseGrepOutput([]byte(input), 10)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Content != `host: "localhost:8080"` {
		t.Errorf("Content = %q, want %q", results[0].Content, `host: "localhost:8080"`)
	}
}
