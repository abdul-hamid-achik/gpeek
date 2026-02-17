package diff

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
		{
			name: "single file add",
			input: `diff --git a/test.go b/test.go
new file mode 100644
--- /dev/null
+++ b/test.go
@@ -0,0 +1,3 @@
+package main
+
+func main() {}`,
			expected: 1,
		},
		{
			name: "single file modify",
			input: `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}`,
			expected: 1,
		},
		{
			name: "multiple files",
			input: `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1 +1 @@
-old
+new
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1 +1 @@
-old2
+new2`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			if len(result.Files) != tt.expected {
				t.Errorf("expected %d files, got %d", tt.expected, len(result.Files))
			}
		})
	}
}

func TestDiffStats(t *testing.T) {
	input := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -1,3 +1,5 @@
 package main

+import "fmt"
+
 func main() {
-	println("hello")
+	fmt.Println("hello")
 }`

	d := Parse(input)
	added, removed := d.Stats()

	if added != 3 {
		t.Errorf("expected 3 additions, got %d", added)
	}
	if removed != 1 {
		t.Errorf("expected 1 removal, got %d", removed)
	}
}

func TestHunkParsing(t *testing.T) {
	input := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -10,5 +10,7 @@ func foo() {
 	bar()
+	baz()
+	qux()
 	end()
 }`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if len(file.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(file.Hunks))
	}

	hunk := file.Hunks[0]
	if hunk.OldStart != 10 {
		t.Errorf("expected old start 10, got %d", hunk.OldStart)
	}
	if hunk.OldCount != 5 {
		t.Errorf("expected old count 5, got %d", hunk.OldCount)
	}
	if hunk.NewStart != 10 {
		t.Errorf("expected new start 10, got %d", hunk.NewStart)
	}
	if hunk.NewCount != 7 {
		t.Errorf("expected new count 7, got %d", hunk.NewCount)
	}
}

func TestNewFile(t *testing.T) {
	input := `diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1 @@
+package new`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	if !d.Files[0].IsNew {
		t.Error("expected file to be marked as new")
	}
}

func TestDeletedFile(t *testing.T) {
	input := `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1 +0,0 @@
-package old`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	if !d.Files[0].IsDelete {
		t.Error("expected file to be marked as deleted")
	}
}

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "normal text",
			content:  "hello world",
			expected: false,
		},
		{
			name:     "text with newlines",
			content:  "line1\nline2\nline3",
			expected: false,
		},
		{
			name:     "unicode text",
			content:  "こんにちは世界",
			expected: false,
		},
		{
			name:     "null byte",
			content:  "hello\x00world",
			expected: true,
		},
		{
			name:     "binary data with null",
			content:  "\x00\x01\x02\x03",
			expected: true,
		},
		{
			name:     "invalid utf8 sequence",
			content:  "\xff\xfe",
			expected: true,
		},
		{
			name:     "mixed valid and invalid utf8",
			content:  "hello\x80world",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryContent(tt.content)
			if result != tt.expected {
				t.Errorf("isBinaryContent(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestBinaryFileDetection(t *testing.T) {
	// Test that git's "Binary files" prefix is detected
	input := `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	if !d.Files[0].IsBinary {
		t.Error("expected file to be marked as binary")
	}
}

func TestBinaryContentInHunk(t *testing.T) {
	// Test that binary content in hunk lines marks file as binary
	input := "diff --git a/checksum b/checksum\n" +
		"--- a/checksum\n" +
		"+++ b/checksum\n" +
		"@@ -1 +1 @@\n" +
		"-old\n" +
		"+new\x00binary"

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	if !d.Files[0].IsBinary {
		t.Error("expected file with binary content to be marked as binary")
	}
}

// --- New tests below ---

func TestRenameDetection(t *testing.T) {
	input := `diff --git a/old_name.go b/new_name.go
similarity index 95%
rename from old_name.go
rename to new_name.go
--- a/old_name.go
+++ b/new_name.go
@@ -1,3 +1,3 @@
 package main
-// old comment
+// new comment
 func main() {}`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if !file.IsRename {
		t.Error("expected file to be marked as rename")
	}
	if file.OldName != "old_name.go" {
		t.Errorf("expected old name 'old_name.go', got %q", file.OldName)
	}
	if file.NewName != "new_name.go" {
		t.Errorf("expected new name 'new_name.go', got %q", file.NewName)
	}
}

func TestMultipleHunksInSingleFile(t *testing.T) {
	input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"

 func foo() {}
@@ -20,3 +21,4 @@
 func bar() {
+	fmt.Println("bar")
 	return
 }`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if len(file.Hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(file.Hunks))
	}

	if file.Hunks[0].OldStart != 1 {
		t.Errorf("first hunk: expected old start 1, got %d", file.Hunks[0].OldStart)
	}
	if file.Hunks[1].OldStart != 20 {
		t.Errorf("second hunk: expected old start 20, got %d", file.Hunks[1].OldStart)
	}
}

func TestLineNumberTracking(t *testing.T) {
	input := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -5,4 +5,5 @@
 line5
-line6_old
+line6_new
+line6b_new
 line7
 line8`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	hunk := d.Files[0].Hunks[0]
	lines := hunk.Lines

	tests := []struct {
		idx       int
		typ       DiffType
		oldNum    int
		newNum    int
		content   string
	}{
		{0, DiffContext, 5, 5, "line5"},
		{1, DiffRemove, 6, 0, "line6_old"},
		{2, DiffAdd, 0, 6, "line6_new"},
		{3, DiffAdd, 0, 7, "line6b_new"},
		{4, DiffContext, 7, 8, "line7"},
		{5, DiffContext, 8, 9, "line8"},
	}

	if len(lines) != len(tests) {
		t.Fatalf("expected %d lines, got %d", len(tests), len(lines))
	}

	for _, tt := range tests {
		line := lines[tt.idx]
		if line.Type != tt.typ {
			t.Errorf("line %d: expected type %v, got %v", tt.idx, tt.typ, line.Type)
		}
		if line.OldNumber != tt.oldNum {
			t.Errorf("line %d: expected old number %d, got %d", tt.idx, tt.oldNum, line.OldNumber)
		}
		if line.NewNumber != tt.newNum {
			t.Errorf("line %d: expected new number %d, got %d", tt.idx, tt.newNum, line.NewNumber)
		}
		if line.Content != tt.content {
			t.Errorf("line %d: expected content %q, got %q", tt.idx, tt.content, line.Content)
		}
	}
}

func TestHunkHeaderWithoutCount(t *testing.T) {
	// When the hunk count is 1, git omits the ",1"
	input := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -1 +1 @@
-old line
+new line`

	d := Parse(input)

	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	hunk := d.Files[0].Hunks[0]
	if hunk.OldCount != 1 {
		t.Errorf("expected old count 1 (default), got %d", hunk.OldCount)
	}
	if hunk.NewCount != 1 {
		t.Errorf("expected new count 1 (default), got %d", hunk.NewCount)
	}
}

func TestHunkHeaderWithFunctionContext(t *testing.T) {
	input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -10,3 +10,4 @@ func myFunction() {
 	existing()
+	added()
 	more()
 }`

	d := Parse(input)
	hunk := d.Files[0].Hunks[0]

	if !strings.Contains(hunk.Header, "func myFunction()") {
		t.Errorf("expected hunk header to contain function context, got %q", hunk.Header)
	}
}

func TestDiffTypeString(t *testing.T) {
	tests := []struct {
		dt       DiffType
		expected string
	}{
		{DiffContext, "context"},
		{DiffAdd, "add"},
		{DiffRemove, "remove"},
		{DiffMeta, "meta"},
		{DiffHunk, "hunk"},
		{DiffBinary, "binary"},
		{DiffType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.expected {
				t.Errorf("DiffType(%d).String() = %q, want %q", tt.dt, got, tt.expected)
			}
		})
	}
}

func TestDiffTypeMarshalJSON(t *testing.T) {
	tests := []struct {
		dt       DiffType
		expected string
	}{
		{DiffContext, `"context"`},
		{DiffAdd, `"add"`},
		{DiffRemove, `"remove"`},
		{DiffBinary, `"binary"`},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got, err := json.Marshal(tt.dt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), tt.expected)
			}
		})
	}
}

func TestFileCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty diff", "", 0},
		{"one file", `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new`, 1},
		{"three files", `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old
+new
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-old
+new
diff --git a/c.go b/c.go
--- a/c.go
+++ b/c.go
@@ -1 +1 @@
-old
+new`, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Parse(tt.input)
			if got := d.FileCount(); got != tt.expected {
				t.Errorf("FileCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestStatsEmptyDiff(t *testing.T) {
	d := Parse("")
	added, removed := d.Stats()
	if added != 0 || removed != 0 {
		t.Errorf("empty diff stats: expected 0/0, got %d/%d", added, removed)
	}
}

func TestMalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"random text", "this is not a diff"},
		{"partial header", "diff --git a/foo"},
		{"hunk without file", "@@ -1,3 +1,3 @@\n foo\n-bar\n+baz"},
		{"only plus lines", "+line1\n+line2\n+line3"},
		{"only minus lines", "-line1\n-line2\n-line3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			d := Parse(tt.input)
			if d == nil {
				t.Error("Parse returned nil")
			}
		})
	}
}

func TestFileNamesExtraction(t *testing.T) {
	input := `diff --git a/path/to/file.go b/path/to/file.go
--- a/path/to/file.go
+++ b/path/to/file.go
@@ -1 +1 @@
-old
+new`

	d := Parse(input)
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if file.OldName != "path/to/file.go" {
		t.Errorf("expected old name 'path/to/file.go', got %q", file.OldName)
	}
	if file.NewName != "path/to/file.go" {
		t.Errorf("expected new name 'path/to/file.go', got %q", file.NewName)
	}
}

func TestDiffWithOnlyBinaryMarker(t *testing.T) {
	// Binary file with no hunks
	input := `diff --git a/logo.png b/logo.png
new file mode 100644
Binary files /dev/null and b/logo.png differ`

	d := Parse(input)
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if !file.IsBinary {
		t.Error("expected binary flag")
	}
	if len(file.Hunks) != 0 {
		t.Errorf("binary file should have no hunks, got %d", len(file.Hunks))
	}
}

func TestDiffWithUnicodeContent(t *testing.T) {
	input := `diff --git a/i18n.go b/i18n.go
--- a/i18n.go
+++ b/i18n.go
@@ -1,3 +1,3 @@
 package i18n
-var greeting = "Hello"
+var greeting = "こんにちは"
 var farewell = "Goodbye"`

	d := Parse(input)
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	added, removed := d.Stats()
	if added != 1 || removed != 1 {
		t.Errorf("expected 1 add 1 remove, got %d/%d", added, removed)
	}

	// Verify the unicode content was preserved
	for _, line := range d.Files[0].Hunks[0].Lines {
		if line.Type == DiffAdd {
			if !strings.Contains(line.Content, "こんにちは") {
				t.Errorf("unicode content not preserved: %q", line.Content)
			}
		}
	}
}

func TestDiffLargeHunkNumbers(t *testing.T) {
	input := `diff --git a/big.go b/big.go
--- a/big.go
+++ b/big.go
@@ -99999,3 +100001,4 @@
 existing
+added
 more
 end`

	d := Parse(input)
	hunk := d.Files[0].Hunks[0]

	if hunk.OldStart != 99999 {
		t.Errorf("expected old start 99999, got %d", hunk.OldStart)
	}
	if hunk.NewStart != 100001 {
		t.Errorf("expected new start 100001, got %d", hunk.NewStart)
	}
}

func TestDiffFileWithNoChanges(t *testing.T) {
	// File header exists but no hunks (mode change only)
	input := `diff --git a/script.sh b/script.sh
old mode 100644
new mode 100755`

	d := Parse(input)
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	file := d.Files[0]
	if len(file.Hunks) != 0 {
		t.Errorf("expected 0 hunks for mode-only change, got %d", len(file.Hunks))
	}
}

func TestDiffMultipleFilesStats(t *testing.T) {
	input := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,2 +1,3 @@
 package a
+// added line
 func A() {}
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1,3 +1,2 @@
 package b
-// removed line
 func B() {}`

	d := Parse(input)
	added, removed := d.Stats()

	if added != 1 {
		t.Errorf("expected 1 addition total, got %d", added)
	}
	if removed != 1 {
		t.Errorf("expected 1 removal total, got %d", removed)
	}
}

func TestParseEmptyLineInHunk(t *testing.T) {
	// Empty lines within a hunk are treated as context lines
	input := "diff --git a/test.go b/test.go\n" +
		"--- a/test.go\n" +
		"+++ b/test.go\n" +
		"@@ -1,4 +1,4 @@\n" +
		" line1\n" +
		"\n" +
		"-old\n" +
		"+new"

	d := Parse(input)
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}

	hunk := d.Files[0].Hunks[0]
	// Should have: context "line1", context "" (empty), remove "old", add "new"
	if len(hunk.Lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(hunk.Lines))
	}

	// The empty line should be context
	if hunk.Lines[1].Type != DiffContext {
		t.Errorf("empty line should be context, got %v", hunk.Lines[1].Type)
	}
}
