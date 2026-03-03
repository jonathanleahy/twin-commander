package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// HighlightMode determines which syntax highlighting rules to apply.
type HighlightMode int

const (
	HighlightNone HighlightMode = iota
	HighlightGo
	HighlightJS
	HighlightPython
	HighlightRust
	HighlightC
	HighlightShell
	HighlightYAML
	HighlightJSON
	HighlightHTML
	HighlightCSS
	HighlightSQL
	HighlightMarkdown
	HighlightDiff
)

// DetectHighlight returns the highlight mode based on file extension.
func DetectHighlight(path string) HighlightMode {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	// Special filenames
	switch base {
	case "makefile", "gnumakefile":
		return HighlightShell
	case "dockerfile":
		return HighlightShell
	}

	switch ext {
	case ".go":
		return HighlightGo
	case ".js", ".jsx", ".ts", ".tsx", ".mjs":
		return HighlightJS
	case ".py":
		return HighlightPython
	case ".rs":
		return HighlightRust
	case ".c", ".h", ".cpp", ".hpp", ".cc", ".cxx", ".cs", ".java", ".kt":
		return HighlightC
	case ".sh", ".bash", ".zsh", ".fish":
		return HighlightShell
	case ".yaml", ".yml", ".toml":
		return HighlightYAML
	case ".json":
		return HighlightJSON
	case ".html", ".htm", ".xml", ".svg":
		return HighlightHTML
	case ".css", ".scss", ".sass", ".less":
		return HighlightCSS
	case ".sql":
		return HighlightSQL
	case ".md", ".markdown":
		return HighlightMarkdown
	case ".diff", ".patch":
		return HighlightDiff
	}
	return HighlightNone
}

// HighlightContent applies tview color tags to source code for the given mode.
// This is a lightweight highlighter — keywords and comments only.
func HighlightContent(content string, mode HighlightMode) string {
	if mode == HighlightNone {
		return tviewEscape(content)
	}

	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		result = append(result, highlightLine(line, mode))
	}

	return strings.Join(result, "\n")
}

// highlightLine applies simple syntax highlighting to a single line.
func highlightLine(line string, mode HighlightMode) string {
	trimmed := strings.TrimSpace(line)

	// Diff highlighting
	if mode == HighlightDiff {
		if strings.HasPrefix(trimmed, "+") && !strings.HasPrefix(trimmed, "+++") {
			return "[green]" + tviewEscape(line) + "[-]"
		}
		if strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "---") {
			return "[red]" + tviewEscape(line) + "[-]"
		}
		if strings.HasPrefix(trimmed, "@@") {
			return "[cyan]" + tviewEscape(line) + "[-]"
		}
		return tviewEscape(line)
	}

	// Comment detection
	switch mode {
	case HighlightGo, HighlightJS, HighlightRust, HighlightC, HighlightCSS, HighlightSQL:
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			return "[gray]" + tviewEscape(line) + "[-]"
		}
	case HighlightPython, HighlightShell, HighlightYAML:
		if strings.HasPrefix(trimmed, "#") {
			return "[gray]" + tviewEscape(line) + "[-]"
		}
	case HighlightHTML:
		if strings.HasPrefix(trimmed, "<!--") {
			return "[gray]" + tviewEscape(line) + "[-]"
		}
	case HighlightMarkdown:
		if strings.HasPrefix(trimmed, "#") {
			return "[yellow::b]" + tviewEscape(line) + "[-::-]"
		}
		if strings.HasPrefix(trimmed, "```") {
			return "[cyan]" + tviewEscape(line) + "[-]"
		}
		if strings.HasPrefix(trimmed, ">") {
			return "[gray]" + tviewEscape(line) + "[-]"
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			return "[green]" + tviewEscape(line) + "[-]"
		}
	}

	// Keyword highlighting for programming languages
	if mode == HighlightGo || mode == HighlightJS || mode == HighlightRust ||
		mode == HighlightC || mode == HighlightPython {
		return highlightKeywords(line, mode)
	}

	// JSON: highlight keys
	if mode == HighlightJSON {
		if strings.Contains(trimmed, ":") {
			return highlightJSONLine(line)
		}
	}

	return tviewEscape(line)
}

// highlightKeywords applies keyword coloring to a line.
func highlightKeywords(line string, mode HighlightMode) string {
	var keywords []string
	switch mode {
	case HighlightGo:
		keywords = []string{"func", "type", "struct", "interface", "package", "import",
			"return", "if", "else", "for", "range", "switch", "case", "default",
			"var", "const", "defer", "go", "chan", "select", "map", "nil", "true", "false"}
	case HighlightJS:
		keywords = []string{"function", "const", "let", "var", "return", "if", "else",
			"for", "while", "class", "import", "export", "from", "async", "await",
			"new", "this", "null", "undefined", "true", "false", "try", "catch"}
	case HighlightPython:
		keywords = []string{"def", "class", "return", "if", "elif", "else", "for",
			"while", "import", "from", "as", "with", "try", "except", "finally",
			"None", "True", "False", "self", "yield", "lambda", "pass", "raise"}
	case HighlightRust:
		keywords = []string{"fn", "let", "mut", "struct", "enum", "impl", "trait",
			"use", "mod", "pub", "return", "if", "else", "for", "while", "match",
			"self", "Self", "true", "false", "None", "Some", "Ok", "Err"}
	case HighlightC:
		keywords = []string{"int", "char", "void", "return", "if", "else", "for",
			"while", "struct", "typedef", "enum", "switch", "case", "break",
			"continue", "const", "static", "class", "public", "private", "protected"}
	}

	escaped := tviewEscape(line)

	// Simple word boundary replacement — only color standalone keywords
	for _, kw := range keywords {
		escaped = highlightWord(escaped, kw, "[blue]", "[-]")
	}

	return escaped
}

// highlightWord colors whole-word occurrences of a keyword in a line.
func highlightWord(line, word, prefix, suffix string) string {
	result := line
	idx := 0
	for {
		pos := strings.Index(result[idx:], word)
		if pos < 0 {
			break
		}
		absPos := idx + pos

		// Check word boundaries
		before := absPos == 0 || !isIdentChar(rune(result[absPos-1]))
		after := absPos+len(word) >= len(result) || !isIdentChar(rune(result[absPos+len(word)]))

		if before && after {
			replacement := prefix + word + suffix
			result = result[:absPos] + replacement + result[absPos+len(word):]
			idx = absPos + len(replacement)
		} else {
			idx = absPos + len(word)
		}
	}
	return result
}

func isIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// highlightJSONLine applies simple JSON key highlighting.
func highlightJSONLine(line string) string {
	escaped := tviewEscape(line)
	// Find quoted keys before ":"
	if idx := strings.Index(escaped, ":"); idx > 0 {
		// Look for the key portion (everything before the colon, including quotes)
		keyPart := escaped[:idx]
		valuePart := escaped[idx:]
		return fmt.Sprintf("[yellow]%s[-]%s", keyPart, valuePart)
	}
	return escaped
}

// tviewEscape escapes tview color tag characters in plain text.
func tviewEscape(s string) string {
	s = strings.ReplaceAll(s, "[", "[[]")
	return s
}
