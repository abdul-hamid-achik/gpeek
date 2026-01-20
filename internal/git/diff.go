package git

import (
	"bytes"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (r *Repository) FileDiff(path string, staged bool) (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}

	status, err := wt.Status()
	if err != nil {
		return "", err
	}

	fileStatus, ok := status[path]
	if !ok {
		return "", fmt.Errorf("file not found in status: %s", path)
	}

	head, err := r.repo.Head()
	if err != nil {
		return r.untrackedFileDiff(path)
	}

	commit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return r.untrackedFileDiff(path)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	if staged {
		return r.stagedDiff(path, tree)
	}

	if fileStatus.Staging == git.Untracked && fileStatus.Worktree == git.Untracked {
		return r.untrackedFileDiff(path)
	}

	return r.workingDiff(path, tree)
}

func (r *Repository) stagedDiff(path string, headTree *object.Tree) (string, error) {
	idx, err := r.repo.Storer.Index()
	if err != nil {
		return "", err
	}

	var oldContent, newContent string

	if f, err := headTree.File(path); err == nil {
		oldContent, _ = f.Contents()
	}

	for _, entry := range idx.Entries {
		if entry.Name == path {
			blob, err := r.repo.BlobObject(entry.Hash)
			if err == nil {
				reader, err := blob.Reader()
				if err == nil {
					buf := new(bytes.Buffer)
					_, _ = buf.ReadFrom(reader)
					newContent = buf.String()
					_ = reader.Close()
				}
			}
			break
		}
	}

	return generateUnifiedDiff(path, oldContent, newContent), nil
}

func (r *Repository) workingDiff(path string, headTree *object.Tree) (string, error) {
	var oldContent string

	// First try to get content from INDEX (staged version)
	// This ensures we compare INDEX → WORKING TREE for unstaged changes
	idx, err := r.repo.Storer.Index()
	if err == nil {
		for _, entry := range idx.Entries {
			if entry.Name == path {
				blob, err := r.repo.BlobObject(entry.Hash)
				if err == nil {
					reader, err := blob.Reader()
					if err == nil {
						buf := new(bytes.Buffer)
						_, _ = buf.ReadFrom(reader)
						oldContent = buf.String()
						_ = reader.Close()
					}
				}
				break
			}
		}
	}

	// If not in INDEX, fall back to HEAD
	if oldContent == "" {
		if f, err := headTree.File(path); err == nil {
			oldContent, _ = f.Contents()
		}
	}

	wt, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}

	fs := wt.Filesystem
	file, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(file)
	newContent := buf.String()

	return generateUnifiedDiff(path, oldContent, newContent), nil
}

func (r *Repository) untrackedFileDiff(path string) (string, error) {
	wt, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}

	fs := wt.Filesystem
	file, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(file)
	content := buf.String()

	return generateUnifiedDiff(path, "", content), nil
}

func (r *Repository) CommitDiff(hash string) (string, error) {
	h := plumbing.NewHash(hash)

	commit, err := r.repo.CommitObject(h)
	if err != nil {
		return "", err
	}

	var parentTree *object.Tree
	if commit.NumParents() > 0 {
		parent, err := commit.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	if parentTree == nil {
		parentTree = &object.Tree{}
	}

	changes, err := parentTree.Diff(commitTree)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}
		result.WriteString(formatPatch(patch))
	}

	return result.String(), nil
}

func generateUnifiedDiff(path, oldContent, newContent string) string {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	var result bytes.Buffer

	result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
	result.WriteString(fmt.Sprintf("--- a/%s\n", path))
	result.WriteString(fmt.Sprintf("+++ b/%s\n", path))

	hunks := computeHunks(oldLines, newLines)
	for _, hunk := range hunks {
		result.WriteString(hunk)
	}

	return result.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// diffOp represents a single diff operation
type diffOp struct {
	kind   int    // opEqual, opAdd, opDelete
	line   string
	oldIdx int // 1-based line number in old file (-1 for additions)
	newIdx int // 1-based line number in new file (-1 for deletions)
}

const (
	opEqual  = 0
	opAdd    = 1
	opDelete = 2
)

func computeHunks(oldLines, newLines []string) []string {
	if len(oldLines) == 0 && len(newLines) == 0 {
		return nil
	}

	if len(oldLines) == 0 {
		var hunk bytes.Buffer
		hunk.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(newLines)))
		for _, line := range newLines {
			hunk.WriteString("+" + line + "\n")
		}
		return []string{hunk.String()}
	}

	if len(newLines) == 0 {
		var hunk bytes.Buffer
		hunk.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(oldLines)))
		for _, line := range oldLines {
			hunk.WriteString("-" + line + "\n")
		}
		return []string{hunk.String()}
	}

	// Compute edit script using LCS
	ops := computeEditScript(oldLines, newLines)

	// Find change regions (sequences of non-equal operations)
	regions := findChangeRegions(ops)

	if len(regions) == 0 {
		return nil // No changes
	}

	// Build hunks with 3 lines of context
	return buildHunksWithContext(ops, regions, 3)
}

// computeEditScript generates a sequence of diff operations using LCS
func computeEditScript(oldLines, newLines []string) []diffOp {
	m, n := len(oldLines), len(newLines)

	// Build LCS table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to produce edit script
	var ops []diffOp
	i, j := m, n

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			ops = append(ops, diffOp{kind: opEqual, line: oldLines[i-1], oldIdx: i, newIdx: j})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			ops = append(ops, diffOp{kind: opAdd, line: newLines[j-1], oldIdx: -1, newIdx: j})
			j--
		} else {
			ops = append(ops, diffOp{kind: opDelete, line: oldLines[i-1], oldIdx: i, newIdx: -1})
			i--
		}
	}

	// Reverse to get forward order
	for left, right := 0, len(ops)-1; left < right; left, right = left+1, right-1 {
		ops[left], ops[right] = ops[right], ops[left]
	}

	return ops
}

// findChangeRegions finds contiguous sequences of non-equal operations
// Returns slice of [start, end) indices into ops
func findChangeRegions(ops []diffOp) [][2]int {
	var regions [][2]int
	inChange := false
	start := 0

	for i, op := range ops {
		if op.kind != opEqual {
			if !inChange {
				start = i
				inChange = true
			}
		} else {
			if inChange {
				regions = append(regions, [2]int{start, i})
				inChange = false
			}
		}
	}

	if inChange {
		regions = append(regions, [2]int{start, len(ops)})
	}

	return regions
}

// buildHunksWithContext creates hunks with context lines, merging nearby regions
func buildHunksWithContext(ops []diffOp, regions [][2]int, contextLines int) []string {
	var hunks []string

	// Merge regions that are close together (within 2*contextLines)
	merged := mergeCloseRegions(regions, ops, contextLines)

	for _, region := range merged {
		hunk := buildSingleHunk(ops, region[0], region[1], contextLines)
		if hunk != "" {
			hunks = append(hunks, hunk)
		}
	}

	return hunks
}

// mergeCloseRegions merges change regions that would have overlapping context
func mergeCloseRegions(regions [][2]int, ops []diffOp, contextLines int) [][2]int {
	if len(regions) == 0 {
		return nil
	}

	merged := [][2]int{regions[0]}

	for i := 1; i < len(regions); i++ {
		last := &merged[len(merged)-1]
		curr := regions[i]

		// Count equal ops between last region end and current region start
		equalCount := 0
		for j := last[1]; j < curr[0]; j++ {
			if ops[j].kind == opEqual {
				equalCount++
			}
		}

		// If gap is small enough, merge the regions
		if equalCount <= 2*contextLines {
			last[1] = curr[1]
		} else {
			merged = append(merged, curr)
		}
	}

	return merged
}

// buildSingleHunk builds a single hunk for a change region with context
func buildSingleHunk(ops []diffOp, changeStart, changeEnd, contextLines int) string {
	// Find context boundaries
	contextStart := changeStart
	contextEnd := changeEnd

	// Add leading context (up to contextLines equal ops before the change)
	leadingContext := 0
	for i := changeStart - 1; i >= 0 && leadingContext < contextLines; i-- {
		if ops[i].kind == opEqual {
			contextStart = i
			leadingContext++
		}
	}

	// Add trailing context (up to contextLines equal ops after the change)
	trailingContext := 0
	for i := changeEnd; i < len(ops) && trailingContext < contextLines; i++ {
		if ops[i].kind == opEqual {
			contextEnd = i + 1
			trailingContext++
		}
	}

	// Calculate line numbers for hunk header
	var oldStart, oldCount, newStart, newCount int

	// Find first old line number and first new line number
	for i := contextStart; i < contextEnd; i++ {
		op := ops[i]
		if oldStart == 0 && op.oldIdx > 0 {
			oldStart = op.oldIdx
		}
		if newStart == 0 && op.newIdx > 0 {
			newStart = op.newIdx
		}
		if oldStart > 0 && newStart > 0 {
			break
		}
	}

	// If still not set, use 1 or 0
	if oldStart == 0 {
		oldStart = 1
	}
	if newStart == 0 {
		newStart = 1
	}

	// Count lines in each file
	for i := contextStart; i < contextEnd; i++ {
		op := ops[i]
		switch op.kind {
		case opEqual:
			oldCount++
			newCount++
		case opDelete:
			oldCount++
		case opAdd:
			newCount++
		}
	}

	// Build the hunk content
	var hunk bytes.Buffer
	hunk.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount))

	for i := contextStart; i < contextEnd; i++ {
		op := ops[i]
		switch op.kind {
		case opEqual:
			hunk.WriteString(" " + op.line + "\n")
		case opDelete:
			hunk.WriteString("-" + op.line + "\n")
		case opAdd:
			hunk.WriteString("+" + op.line + "\n")
		}
	}

	return hunk.String()
}

func formatPatch(patch *object.Patch) string {
	var buf bytes.Buffer
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		var fromPath, toPath string
		if from != nil {
			fromPath = from.Path()
		}
		if to != nil {
			toPath = to.Path()
		}
		if fromPath == "" {
			fromPath = toPath
		}
		if toPath == "" {
			toPath = fromPath
		}

		buf.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fromPath, toPath))
		buf.WriteString(fmt.Sprintf("--- a/%s\n", fromPath))
		buf.WriteString(fmt.Sprintf("+++ b/%s\n", toPath))

		// Calculate line counts for hunk header
		var oldCount, newCount int
		for _, chunk := range fp.Chunks() {
			lines := splitLines(chunk.Content())
			switch chunk.Type() {
			case diff.Add:
				newCount += len(lines)
			case diff.Delete:
				oldCount += len(lines)
			case diff.Equal:
				oldCount += len(lines)
				newCount += len(lines)
			}
		}

		// Write hunk header
		if oldCount > 0 || newCount > 0 {
			fmt.Fprintf(&buf, "@@ -%d,%d +%d,%d @@\n", 1, oldCount, 1, newCount)
		}

		// Write chunk content
		for _, chunk := range fp.Chunks() {
			switch chunk.Type() {
			case diff.Add:
				for _, line := range splitLines(chunk.Content()) {
					buf.WriteString("+" + line + "\n")
				}
			case diff.Delete:
				for _, line := range splitLines(chunk.Content()) {
					buf.WriteString("-" + line + "\n")
				}
			case diff.Equal:
				for _, line := range splitLines(chunk.Content()) {
					buf.WriteString(" " + line + "\n")
				}
			}
		}
	}
	return buf.String()
}
