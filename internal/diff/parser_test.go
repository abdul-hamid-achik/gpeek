package diff

import (
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
