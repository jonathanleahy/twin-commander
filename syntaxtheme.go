package main

// SyntaxTheme defines colors for syntax highlighting in the preview pane.
type SyntaxTheme struct {
	Name    string
	Keyword string // tview color tag for keywords
	Comment string // tview color tag for comments
	String  string // tview color tag for string literals
	Heading string // tview color tag for markdown headings
	Added   string // tview color tag for diff additions
	Removed string // tview color tag for diff deletions
	Meta    string // tview color tag for diff meta (@@)
	JSONKey string // tview color tag for JSON keys
	List    string // tview color tag for markdown list items
}

// SyntaxThemeNames returns the ordered list of available syntax theme names.
func SyntaxThemeNames() []string {
	return []string{"Default", "Monokai", "Dracula", "Solarized", "GitHub", "Nord"}
}

var syntaxThemes = map[string]SyntaxTheme{
	"Default": {
		Name:    "Default",
		Keyword: "blue",
		Comment: "gray",
		String:  "green",
		Heading: "yellow::b",
		Added:   "green",
		Removed: "red",
		Meta:    "cyan",
		JSONKey: "yellow",
		List:    "green",
	},
	"Monokai": {
		Name:    "Monokai",
		Keyword: "red",
		Comment: "gray",
		String:  "yellow",
		Heading: "orange::b",
		Added:   "green",
		Removed: "red",
		Meta:    "purple",
		JSONKey: "purple",
		List:    "yellow",
	},
	"Dracula": {
		Name:    "Dracula",
		Keyword: "purple",
		Comment: "gray",
		String:  "yellow",
		Heading: "pink::b",
		Added:   "green",
		Removed: "red",
		Meta:    "cyan",
		JSONKey: "cyan",
		List:    "green",
	},
	"Solarized": {
		Name:    "Solarized",
		Keyword: "green",
		Comment: "gray",
		String:  "cyan",
		Heading: "orange::b",
		Added:   "green",
		Removed: "red",
		Meta:    "blue",
		JSONKey: "blue",
		List:    "cyan",
	},
	"GitHub": {
		Name:    "GitHub",
		Keyword: "purple",
		Comment: "gray",
		String:  "blue",
		Heading: "blue::b",
		Added:   "green",
		Removed: "red",
		Meta:    "cyan",
		JSONKey: "teal",
		List:    "blue",
	},
	"Nord": {
		Name:    "Nord",
		Keyword: "blue",
		Comment: "gray",
		String:  "green",
		Heading: "yellow::b",
		Added:   "green",
		Removed: "red",
		Meta:    "cyan",
		JSONKey: "teal",
		List:    "aqua",
	},
}

// GetSyntaxTheme returns the named syntax theme, or Default if not found.
func GetSyntaxTheme(name string) SyntaxTheme {
	if t, ok := syntaxThemes[name]; ok {
		return t
	}
	return syntaxThemes["Default"]
}
