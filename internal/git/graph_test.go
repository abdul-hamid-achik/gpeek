package git

import (
	"testing"
	"time"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g == nil {
		t.Fatal("NewGraph returned nil")
	}

	if g.Columns == nil {
		t.Error("Columns map should be initialized")
	}

	if len(g.Nodes) != 0 {
		t.Error("Nodes should be empty initially")
	}
}

func TestGraphAddCommit(t *testing.T) {
	g := NewGraph()

	commit := Commit{
		Hash:    "abc123",
		Message: "Test commit",
		Time:    time.Now(),
		Parents: []string{},
	}

	g.AddCommit(commit)

	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.Nodes))
	}

	if g.Nodes[0].Hash != "abc123" {
		t.Errorf("expected hash 'abc123', got '%s'", g.Nodes[0].Hash)
	}
}

func TestGraphLinearHistory(t *testing.T) {
	g := NewGraph()

	commits := []Commit{
		{Hash: "commit3", Parents: []string{"commit2"}},
		{Hash: "commit2", Parents: []string{"commit1"}},
		{Hash: "commit1", Parents: []string{}},
	}

	for _, c := range commits {
		g.AddCommit(c)
	}

	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}

	for i, node := range g.Nodes {
		if node.Column != 1 {
			t.Errorf("commit %d: expected column 1, got %d", i, node.Column)
		}
	}
}

func TestGraphMergeCommit(t *testing.T) {
	g := NewGraph()

	commit := Commit{
		Hash:    "merge",
		Message: "Merge branch",
		Parents: []string{"parent1", "parent2"},
		IsMerge: true,
	}

	g.AddCommit(commit)

	if !g.Nodes[0].IsMerge {
		t.Error("expected commit to be marked as merge")
	}
}

func TestRenderCommitGraph(t *testing.T) {
	commits := []Commit{
		{Hash: "c3", Parents: []string{"c2"}},
		{Hash: "c2", Parents: []string{"c1"}},
		{Hash: "c1", Parents: []string{}},
	}

	lines := RenderCommitGraph(commits, 40)

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if line == "" {
			t.Errorf("line %d should not be empty", i)
		}
	}
}

func TestGraphRender(t *testing.T) {
	g := NewGraph()

	g.AddCommit(Commit{Hash: "c1", Parents: []string{}})
	g.AddCommit(Commit{Hash: "c2", Parents: []string{"c1"}})

	rendered := g.Render(0, 20)
	if rendered == "" {
		t.Error("render should not return empty string")
	}

	rendered2 := g.Render(1, 20)
	if rendered2 == "" {
		t.Error("render should not return empty string for second commit")
	}
}

func TestGraphRenderOutOfBounds(t *testing.T) {
	g := NewGraph()
	g.AddCommit(Commit{Hash: "c1", Parents: []string{}})

	rendered := g.Render(100, 20)
	if rendered != "" {
		t.Error("render should return empty string for out of bounds index")
	}
}
