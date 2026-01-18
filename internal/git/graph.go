package git

import (
	"strings"
)

type GraphNode struct {
	Hash     string
	Column   int
	Parents  []string
	IsMerge  bool
	Branches []string
}

type Graph struct {
	Nodes   []GraphNode
	Columns map[string]int
	MaxCol  int
}

func NewGraph() *Graph {
	return &Graph{
		Columns: make(map[string]int),
	}
}

func (g *Graph) AddCommit(commit Commit) {
	col := g.findColumn(commit.Hash, commit.Parents)

	node := GraphNode{
		Hash:    commit.Hash,
		Column:  col,
		Parents: commit.Parents,
		IsMerge: len(commit.Parents) > 1,
	}

	g.Nodes = append(g.Nodes, node)
	g.Columns[commit.Hash] = col

	for i, parent := range commit.Parents {
		if _, exists := g.Columns[parent]; !exists {
			if i == 0 {
				g.Columns[parent] = col
			} else {
				g.MaxCol++
				g.Columns[parent] = g.MaxCol
			}
		}
	}
}

func (g *Graph) findColumn(hash string, parents []string) int {
	if col, exists := g.Columns[hash]; exists {
		return col
	}

	for _, node := range g.Nodes {
		for _, parent := range node.Parents {
			if parent == hash {
				return node.Column
			}
		}
	}

	g.MaxCol++
	return g.MaxCol
}

func (g *Graph) Render(nodeIdx int, width int) string {
	if nodeIdx >= len(g.Nodes) {
		return ""
	}

	node := g.Nodes[nodeIdx]
	cols := g.MaxCol + 1

	if cols > width/2 {
		cols = width / 2
	}

	line := make([]rune, cols*2)
	for i := range line {
		line[i] = ' '
	}

	for i := 0; i < cols; i++ {
		if g.isActiveColumn(nodeIdx, i) {
			line[i*2] = '│'
		}
	}

	if node.Column*2 < len(line) {
		if node.IsMerge {
			line[node.Column*2] = '●'
		} else {
			line[node.Column*2] = '○'
		}
	}

	if node.IsMerge && len(node.Parents) > 1 {
		parentCol := g.Columns[node.Parents[1]]
		if parentCol > node.Column {
			for i := node.Column + 1; i < parentCol && i*2 < len(line); i++ {
				if line[i*2] == '│' {
					line[i*2] = '┼'
				} else {
					line[i*2] = '─'
				}
			}
			if parentCol*2-1 < len(line) {
				line[parentCol*2-1] = '╮'
			}
		} else if parentCol < node.Column {
			for i := parentCol + 1; i < node.Column && i*2 < len(line); i++ {
				if line[i*2] == '│' {
					line[i*2] = '┼'
				} else {
					line[i*2] = '─'
				}
			}
			if parentCol*2+1 < len(line) {
				line[parentCol*2+1] = '╭'
			}
		}
	}

	return strings.TrimRight(string(line), " ")
}

func (g *Graph) isActiveColumn(nodeIdx, col int) bool {
	for i := 0; i <= nodeIdx; i++ {
		node := g.Nodes[i]
		if node.Column == col {
			for j := i; j < len(g.Nodes); j++ {
				for _, parent := range g.Nodes[j].Parents {
					if g.Columns[parent] == col {
						if j >= nodeIdx {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func RenderCommitGraph(commits []Commit, width int) []string {
	g := NewGraph()

	for _, c := range commits {
		g.AddCommit(c)
	}

	var lines []string
	for i := range commits {
		lines = append(lines, g.Render(i, width))
	}

	return lines
}
