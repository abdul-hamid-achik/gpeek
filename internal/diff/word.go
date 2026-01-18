package diff

import (
	"strings"
	"unicode"
)

type WordChange struct {
	Text      string
	IsChanged bool
}

type WordDiff struct {
	OldWords []WordChange
	NewWords []WordChange
}

func ComputeWordDiff(oldLine, newLine string) WordDiff {
	oldWords := tokenize(oldLine)
	newWords := tokenize(newLine)

	lcs := longestCommonSubsequence(oldWords, newWords)

	result := WordDiff{}

	oldIdx := 0
	newIdx := 0
	lcsIdx := 0

	for oldIdx < len(oldWords) || newIdx < len(newWords) {
		if lcsIdx < len(lcs) {
			for oldIdx < len(oldWords) && oldWords[oldIdx] != lcs[lcsIdx] {
				result.OldWords = append(result.OldWords, WordChange{
					Text:      oldWords[oldIdx],
					IsChanged: true,
				})
				oldIdx++
			}

			for newIdx < len(newWords) && newWords[newIdx] != lcs[lcsIdx] {
				result.NewWords = append(result.NewWords, WordChange{
					Text:      newWords[newIdx],
					IsChanged: true,
				})
				newIdx++
			}

			if oldIdx < len(oldWords) && newIdx < len(newWords) {
				result.OldWords = append(result.OldWords, WordChange{
					Text:      oldWords[oldIdx],
					IsChanged: false,
				})
				result.NewWords = append(result.NewWords, WordChange{
					Text:      newWords[newIdx],
					IsChanged: false,
				})
				oldIdx++
				newIdx++
				lcsIdx++
			}
		} else {
			for oldIdx < len(oldWords) {
				result.OldWords = append(result.OldWords, WordChange{
					Text:      oldWords[oldIdx],
					IsChanged: true,
				})
				oldIdx++
			}
			for newIdx < len(newWords) {
				result.NewWords = append(result.NewWords, WordChange{
					Text:      newWords[newIdx],
					IsChanged: true,
				})
				newIdx++
			}
		}
	}

	return result
}

func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	var lastType tokenType

	for _, r := range s {
		currentType := getTokenType(r)

		if lastType != currentType && current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}

		current.WriteRune(r)
		lastType = currentType
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

type tokenType int

const (
	tokenWord tokenType = iota
	tokenSpace
	tokenPunct
	tokenNumber
)

func getTokenType(r rune) tokenType {
	switch {
	case unicode.IsSpace(r):
		return tokenSpace
	case unicode.IsLetter(r) || r == '_':
		return tokenWord
	case unicode.IsDigit(r):
		return tokenNumber
	default:
		return tokenPunct
	}
}

func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return nil
	}

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (wd WordDiff) RenderOld(changedStyle, normalStyle func(string) string) string {
	var result strings.Builder
	for _, w := range wd.OldWords {
		if w.IsChanged {
			result.WriteString(changedStyle(w.Text))
		} else {
			result.WriteString(normalStyle(w.Text))
		}
	}
	return result.String()
}

func (wd WordDiff) RenderNew(changedStyle, normalStyle func(string) string) string {
	var result strings.Builder
	for _, w := range wd.NewWords {
		if w.IsChanged {
			result.WriteString(changedStyle(w.Text))
		} else {
			result.WriteString(normalStyle(w.Text))
		}
	}
	return result.String()
}
