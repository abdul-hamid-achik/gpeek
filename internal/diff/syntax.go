package diff

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type SyntaxHighlighter struct {
	style *chroma.Style
}

func NewSyntaxHighlighter(styleName string) *SyntaxHighlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	return &SyntaxHighlighter{style: style}
}

func (h *SyntaxHighlighter) Highlight(code, filename string) ([]HighlightedLine, error) {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	var lines []HighlightedLine
	var currentLine HighlightedLine

	for token := iterator(); token != chroma.EOF; token = iterator() {
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, currentLine)
				currentLine = HighlightedLine{}
			}
			if part != "" {
				currentLine.Tokens = append(currentLine.Tokens, Token{
					Type:  token.Type,
					Value: part,
				})
			}
		}
	}

	if len(currentLine.Tokens) > 0 || len(lines) == 0 {
		lines = append(lines, currentLine)
	}

	return lines, nil
}

func (h *SyntaxHighlighter) HighlightDiff(diff *Diff) (*HighlightedDiff, error) {
	result := &HighlightedDiff{
		Files: make([]HighlightedFileDiff, len(diff.Files)),
	}

	for i, file := range diff.Files {
		hFile := HighlightedFileDiff{
			OldName:  file.OldName,
			NewName:  file.NewName,
			IsBinary: file.IsBinary,
			IsNew:    file.IsNew,
			IsDelete: file.IsDelete,
			IsRename: file.IsRename,
			Hunks:    make([]HighlightedHunk, len(file.Hunks)),
		}

		filename := file.NewName
		if filename == "" || filename == "/dev/null" {
			filename = file.OldName
		}

		for j, hunk := range file.Hunks {
			hHunk := HighlightedHunk{
				OldStart: hunk.OldStart,
				OldCount: hunk.OldCount,
				NewStart: hunk.NewStart,
				NewCount: hunk.NewCount,
				Header:   hunk.Header,
				Lines:    make([]HighlightedLine, len(hunk.Lines)),
			}

			for k, line := range hunk.Lines {
				tokens, _ := h.tokenizeLine(line.Content, filename)
				hHunk.Lines[k] = HighlightedLine{
					Type:      line.Type,
					OldNumber: line.OldNumber,
					NewNumber: line.NewNumber,
					Tokens:    tokens,
				}
			}

			hFile.Hunks[j] = hHunk
		}

		result.Files[i] = hFile
	}

	return result, nil
}

func (h *SyntaxHighlighter) tokenizeLine(content, filename string) ([]Token, error) {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return []Token{{Type: chroma.Text, Value: content}}, nil
	}

	var tokens []Token
	for token := iterator(); token != chroma.EOF; token = iterator() {
		if !strings.Contains(token.Value, "\n") {
			tokens = append(tokens, Token{
				Type:  token.Type,
				Value: token.Value,
			})
		}
	}

	return tokens, nil
}

func (h *SyntaxHighlighter) GetTokenColor(tokenType chroma.TokenType) string {
	entry := h.style.Get(tokenType)
	if entry.Colour.IsSet() {
		return entry.Colour.String()
	}
	return ""
}

type Token struct {
	Type  chroma.TokenType
	Value string
}

type HighlightedLine struct {
	Type      DiffType
	OldNumber int
	NewNumber int
	Tokens    []Token
}

type HighlightedHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []HighlightedLine
}

type HighlightedFileDiff struct {
	OldName  string
	NewName  string
	IsBinary bool
	IsNew    bool
	IsDelete bool
	IsRename bool
	Hunks    []HighlightedHunk
}

type HighlightedDiff struct {
	Files []HighlightedFileDiff
}

func DetectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".jsx":
		return "jsx"
	case ".tsx":
		return "tsx"
	case ".py":
		return "python"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".sh", ".bash":
		return "bash"
	case ".zsh":
		return "zsh"
	case ".ps1":
		return "powershell"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".less":
		return "less"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".md", ".markdown":
		return "markdown"
	case ".toml":
		return "toml"
	case ".ini", ".cfg":
		return "ini"
	case ".dockerfile":
		return "docker"
	case ".makefile":
		return "make"
	default:
		return ""
	}
}
