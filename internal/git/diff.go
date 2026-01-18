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

	if f, err := headTree.File(path); err == nil {
		oldContent, _ = f.Contents()
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

func computeHunks(oldLines, newLines []string) []string {
	var hunks []string

	if len(oldLines) == 0 && len(newLines) == 0 {
		return hunks
	}

	if len(oldLines) == 0 {
		var hunk bytes.Buffer
		hunk.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(newLines)))
		for _, line := range newLines {
			hunk.WriteString("+" + line + "\n")
		}
		hunks = append(hunks, hunk.String())
		return hunks
	}

	if len(newLines) == 0 {
		var hunk bytes.Buffer
		hunk.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(oldLines)))
		for _, line := range oldLines {
			hunk.WriteString("-" + line + "\n")
		}
		hunks = append(hunks, hunk.String())
		return hunks
	}

	lcs := longestCommonSubsequence(oldLines, newLines)
	_ = lcs

	var hunk bytes.Buffer
	hunk.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines)))

	oldIdx, newIdx := 0, 0
	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if oldIdx < len(oldLines) && newIdx < len(newLines) && oldLines[oldIdx] == newLines[newIdx] {
			hunk.WriteString(" " + oldLines[oldIdx] + "\n")
			oldIdx++
			newIdx++
		} else if newIdx < len(newLines) && (oldIdx >= len(oldLines) || !contains(oldLines[oldIdx:], newLines[newIdx])) {
			hunk.WriteString("+" + newLines[newIdx] + "\n")
			newIdx++
		} else if oldIdx < len(oldLines) {
			hunk.WriteString("-" + oldLines[oldIdx] + "\n")
			oldIdx++
		}
	}

	hunks = append(hunks, hunk.String())
	return hunks
}

func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	var lcs []string
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
