package panels

// CursorPosition returns the 1-based cursor position in the files panel.
func (p *FilesPanel) CursorPosition() int {
	if p.totalItems() == 0 {
		return 0
	}
	return p.cursor + 1
}

// TotalItems returns the total number of files displayed in the panel.
func (p *FilesPanel) TotalItems() int {
	return p.totalItems()
}

// CursorPosition returns the 1-based cursor position in the branches panel.
func (p *BranchesPanel) CursorPosition() int {
	if len(p.filteredBranches) == 0 {
		return 0
	}
	return p.cursor + 1
}

// TotalItems returns the total number of branches displayed in the panel.
func (p *BranchesPanel) TotalItems() int {
	return len(p.filteredBranches)
}

// CursorPosition returns the 1-based cursor position in the commits panel.
func (p *CommitsPanel) CursorPosition() int {
	if len(p.filteredCommits) == 0 {
		return 0
	}
	return p.cursor + 1
}

// TotalItems returns the total number of commits displayed in the panel.
func (p *CommitsPanel) TotalItems() int {
	return len(p.filteredCommits)
}
