package modals

import (
	"testing"

	"github.com/abdul-hamid-achik/gpeek/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func TestDiffModalFilePositionTracking(t *testing.T) {
	// Create a simple theme for testing
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	// Create sample diff content with multiple files
	diffContent := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1,2 +1,3 @@
 package utils

+func Helper() {}
diff --git a/file3.go b/file3.go
--- a/file3.go
+++ b/file3.go
@@ -1 +1 @@
-old
+new`

	// Create modal with fixed dimensions
	modal := NewDiffModal(styles, "Test Commit", diffContent, 80, 24)

	// Test 1: Verify file positions are tracked
	if len(modal.filePositions) != 3 {
		t.Errorf("Expected 3 file positions, got %d", len(modal.filePositions))
	}

	// Test 2: Verify initial state - all files collapsed
	for i, expanded := range modal.expanded {
		if expanded {
			t.Errorf("File %d should be collapsed initially", i)
		}
	}

	// Test 3: Verify file start positions
	// File 0 should start at line 0
	if modal.filePositions[0].startLine != 0 {
		t.Errorf("File 0 should start at line 0, got %d", modal.filePositions[0].startLine)
	}

	// File 1 should start at line 1 (after file 0 header)
	if modal.filePositions[1].startLine != 1 {
		t.Errorf("File 1 should start at line 1, got %d", modal.filePositions[1].startLine)
	}

	// File 2 should start at line 2 (after file 0 and 1 headers)
	if modal.filePositions[2].startLine != 2 {
		t.Errorf("File 2 should start at line 2, got %d", modal.filePositions[2].startLine)
	}
}

func TestDiffModalFileExpansion(t *testing.T) {
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	diffContent := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}`

	modal := NewDiffModal(styles, "Test", diffContent, 80, 24)

	// Test expansion
	modal.expanded[0] = true
	modal.renderContent()

	// Verify file position is updated after expansion
	if !modal.filePositions[0].expanded {
		t.Error("File position should track expanded state")
	}

	// Test that endLine is greater than startLine after expansion
	if modal.filePositions[0].endLine <= modal.filePositions[0].startLine {
		t.Error("Expanded file should have endLine > startLine")
	}
}

func TestDiffModalGetVisibleFileIndex(t *testing.T) {
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	// Create diff with multiple files
	diffContent := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1 +1 @@
-old1
+new1
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1 +1 @@
-old2
+new2
diff --git a/file3.go b/file3.go
--- a/file3.go
+++ b/file3.go
@@ -1 +1 @@
-old3
+new3`

	modal := NewDiffModal(styles, "Test", diffContent, 80, 24)

	// When files are collapsed, each takes 1 line
	// File positions: 0->(0,1), 1->(1,2), 2->(2,3)
	// Viewport height is 24, so middle is at line 12
	// With YOffset=0, middle=12, which is past all headers, returns last file

	// Test that getVisibleFileIndex returns a valid file index
	modal.viewport.SetYOffset(0)
	visibleIdx := modal.getVisibleFileIndex()
	if visibleIdx < 0 || visibleIdx >= len(modal.parsedDiff.Files) {
		t.Errorf("getVisibleFileIndex returned invalid index: %d", visibleIdx)
	}

	// Test with expanded files - expand file 1 which has content
	modal.expanded[1] = true
	modal.renderContent()

	// Now file 1 should have more lines
	if modal.filePositions[1].endLine <= modal.filePositions[1].startLine {
		t.Error("Expanded file should have endLine > startLine")
	}
}

func TestDiffModalScrollToFile(t *testing.T) {
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	// Create diff with enough content to require scrolling
	diffContent := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,10 +1,12 @@
 package main

+import "fmt"
 func main() {
-	println("hello")
+	fmt.Println("hello")
+	fmt.Println("world")
 }
-
+// end of file1
 func helper() {}
 func helper2() {}
 func helper3() {}
 func helper4() {}
 func helper5() {}
 func helper6() {}
 func helper7() {}
 func helper8() {}
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1,10 +1,11 @@
 package utils

+func Helper() {}
 func helper1() {}
 func helper2() {}
 func helper3() {}
 func helper4() {}
 func helper5() {}
 func helper6() {}
 func helper7() {}
 func helper8() {}
 func helper9() {}
-func helper10() {}
+func Helper10() {}`

	modal := NewDiffModal(styles, "Test", diffContent, 80, 24)

	// Verify we have 2 files
	if len(modal.parsedDiff.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(modal.parsedDiff.Files))
	}

	// Expand both files so there's content to scroll through
	modal.expanded[0] = true
	modal.expanded[1] = true
	modal.renderContent()

	// Verify file positions are set after expansion
	if len(modal.filePositions) != 2 {
		t.Fatalf("Expected 2 file positions, got %d", len(modal.filePositions))
	}

	// After expansion, files should have substantial content
	// File 0: start=0, end should be > start
	// File 1: start should be >= file 0's end
	if modal.filePositions[0].endLine <= modal.filePositions[0].startLine {
		t.Fatal("File 0 should have content after expansion")
	}
	if modal.filePositions[1].startLine < modal.filePositions[0].endLine {
		t.Fatalf("File 1 should start at or after file 0 ends. File 0 ends at %d, File 1 starts at %d",
			modal.filePositions[0].endLine, modal.filePositions[1].startLine)
	}

	// Test scrolling to file 1
	modal.scrollToFile(1)
	// After scroll, YOffset should be at file 1's start line
	if modal.viewport.YOffset != modal.filePositions[1].startLine {
		t.Errorf("After scrollToFile(1), YOffset should be %d (file 1 start), got %d",
			modal.filePositions[1].startLine, modal.viewport.YOffset)
	}

	// Test scrolling to file 0
	modal.scrollToFile(0)
	if modal.viewport.YOffset != modal.filePositions[0].startLine {
		t.Errorf("After scrollToFile(0), YOffset should be %d (file 0 start), got %d",
			modal.filePositions[0].startLine, modal.viewport.YOffset)
	}

	// Test invalid index (should not panic)
	modal.scrollToFile(-1)
	modal.scrollToFile(100)
}

func TestDiffModalIsAllCollapsed(t *testing.T) {
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	diffContent := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1 +1 @@
-old
+new`

	modal := NewDiffModal(styles, "Test", diffContent, 80, 24)

	// Initially all collapsed
	if !modal.isAllCollapsed() {
		t.Error("Initially all files should be collapsed")
	}

	// Expand file
	modal.expanded[0] = true
	if modal.isAllCollapsed() {
		t.Error("After expanding a file, isAllCollapsed should return false")
	}
}

func TestDiffModalCountFileChanges(t *testing.T) {
	theme := ui.Theme{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#f5c2e7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Subtle:     "#313244",
	}
	styles := ui.NewStyles(theme)

	diffContent := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
+import "os"
 func main() {}
-
+// end`

	modal := NewDiffModal(styles, "Test", diffContent, 80, 24)

	if len(modal.parsedDiff.Files) == 0 {
		t.Fatal("Should have at least one file")
	}

	file := modal.parsedDiff.Files[0]
	adds, dels := modal.countFileChanges(file)

	// The diff has:
	// +import "fmt" (addition)
	// +import "os" (addition)
	// func main() {} (context - no change)
	// - (deletion of empty line)
	// +// end (addition)
	// So: 3 additions, 1 deletion
	if adds != 3 {
		t.Errorf("Expected 3 additions, got %d", adds)
	}
	if dels != 1 {
		t.Errorf("Expected 1 deletion, got %d", dels)
	}
}

// Helper function to set up lipgloss for tests
func init() {
	// Set a default renderer for lipgloss in tests
	lipgloss.SetColorProfile(0) // Use ASCII profile for consistent test output
}
