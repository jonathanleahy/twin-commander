# Twin Commander

A Norton Commander / Midnight Commander style dual-pane terminal file explorer built in Go. Keyboard-driven navigation with real-time search/filter, syntax-highlighted file preview, configurable themes, and cross-platform terminal support.

## Quick Start

```bash
go build -o twin-commander .
./twin-commander
```

Press `q` to exit.

## Installation

### Prerequisites

- Go 1.22 or later

### Build

```bash
go build -o twin-commander .
```

### Verify

```bash
./twin-commander
# Two panels should appear side by side showing the current directory
# Press q to exit
```

## Features

- **Hybrid Tree + Dual-Pane modes** — toggle with Ctrl+T
- **Inline file preview** with syntax highlighting for 14+ languages
- **Fullscreen viewer** with progressive disclosure (preview → viewer → close)
- **Menu bar** with keyboard hotkeys (Alt+F/V/S/G/T/O) and vim-style navigation
- **File operations** — copy, move, rename, delete (with trash support), mkdir
- **Recursive filename search** with real-time results (Ctrl+F)
- **Type-to-filter** with case-insensitive substring matching
- **Git integration** — branch display, file status colors, diff viewer, stage/unstage
- **Directory bookmarks** — save, jump (1-9 keys), manage via Ctrl+B dialog
- **6 color themes** — Default, Dark, Light, Solarized, Monokai, Nord
- **Vim-style keys** — j/k/h/l navigation, gg/G jumps, dd delete, yy yank, p paste
- **Resizable panes** — Alt+arrows to adjust splits (15-85% range)
- **Nerd Font icons** for 60+ file types and directories
- **Beyond Compare integration** — compare files across panels with `b`
- **$EDITOR integration** — open files in your editor with `e`
- **Clipboard support** — copy paths with Alt+C (xclip/xsel/wl-copy)
- **Configurable** — persistent JSON config for all preferences

## Usage

Twin Commander is a full-screen TUI application with no command-line flags or arguments. Launch it from any directory:

```bash
./twin-commander
```

The default view is **hybrid tree mode**: a directory tree on the left and a file panel on the right. Press `Ctrl+T` to switch to classic dual-pane mode.

### Keyboard Reference

#### Navigation

| Key | Action |
|-----|--------|
| j / Down | Move cursor down |
| k / Up | Move cursor up |
| h / Backspace | Navigate up / collapse tree node |
| l / Enter | Navigate into directory / open file preview |
| gg | Jump to top |
| G | Jump to bottom |
| Tab | Cycle active pane forward |
| Shift+Tab | Cycle active pane backward |

#### View

| Key | Action |
|-----|--------|
| Ctrl+T | Toggle hybrid tree / dual-pane mode |
| t | Toggle inline preview pane |
| . | Toggle hidden files |
| s | Cycle sort mode (name / size / date / extension) |
| S | Toggle sort order (ascending / descending) |
| r | Refresh current directory |

#### Search & Filter

| Key | Action |
|-----|--------|
| / | Enter filter mode (type-to-filter) |
| Ctrl+F / F3 | Recursive filename search |

#### File Operations

| Key | Action |
|-----|--------|
| F5 / c | Copy selected to other pane |
| F6 / m | Move selected to other pane |
| F7 / n | Create new directory |
| F8 / dd | Delete (trash or permanent) |
| F2 / R | Rename |
| yy | Yank (mark for copy) |
| p | Paste yanked files |

#### Tools

| Key | Action |
|-----|--------|
| e | Open in $EDITOR |
| b | Launch Beyond Compare |
| Ctrl+G | Git diff for selected file |
| gs | Git stage/unstage |
| Alt+C | Copy path to clipboard |

#### Resize

| Key | Action |
|-----|--------|
| Alt+Left/Right | Adjust horizontal split (5% per press) |
| Alt+Up/Down | Adjust vertical split (file list vs preview) |

#### Menu & System

| Key | Action |
|-----|--------|
| Alt+F/V/S/G/T/O | Open menu by hotkey |
| F9 | Open menu bar |
| Ctrl+B | Open bookmarks dialog |
| 1-9 | Jump to bookmark by number |
| q | Quit |
| Ctrl+C | Force quit |
| Esc | Close overlay / cancel |

#### Preview/Viewer Keys (when preview or viewer is focused)

| Key | Action |
|-----|--------|
| j / Down | Scroll down one line |
| k / Up | Scroll up one line |
| g | Jump to top |
| G | Jump to bottom |
| PgUp | Page up |
| PgDn | Page down |
| t | Return from fullscreen viewer to inline preview |
| Esc | Close viewer/preview |

### Navigation

Navigate into a directory by selecting it and pressing Enter. Go back to the parent directory with Backspace. After going up, the cursor lands on the directory you came from.

### Tab Cycling

Tab cycles focus through all visible panes:

- **Hybrid mode** (tree + files + preview): tree -> files -> preview -> tree
- **Dual mode** (left + right + preview): left -> right -> preview -> left

When the preview pane is focused, it accepts scroll keys (j/k/g/G/PgUp/PgDn) for navigating file content.

### Preview Flow

File preview follows a progressive disclosure model:

1. **Enter on a file** opens an inline preview pane alongside the file list
2. **Enter again** (while preview is showing) escalates to a fullscreen viewer
3. **Escape from fullscreen** returns to the inline preview
4. **Escape from preview** closes the preview pane entirely

Both the inline preview and fullscreen viewer support syntax highlighting and scrollbars.

### Syntax Highlighting

File preview and the fullscreen viewer apply syntax highlighting based on file type:

- **Go** (`.go`): keywords, comments, strings
- **JavaScript/TypeScript** (`.js`, `.ts`, `.jsx`, `.tsx`): keywords, comments
- **Python** (`.py`): keywords, comments
- **Rust** (`.rs`): keywords, comments
- **C/C++/Java** (`.c`, `.cpp`, `.java`): keywords, comments
- **Shell** (`.sh`, `.bash`): comments
- **YAML/TOML** (`.yaml`, `.toml`): comments
- **JSON** (`.json`): key highlighting
- **HTML/CSS** (`.html`, `.css`): comments, tags
- **SQL** (`.sql`): comments
- **Markdown** (`.md`): headers, code blocks, lists, blockquotes
- **Diff** (`.diff`, `.patch`): additions (green), deletions (red), hunks (cyan)

Highlighting is applied in both the inline preview pane and the fullscreen viewer.

### Scrollbars

Preview panes and the fullscreen viewer display scrollbars when content exceeds the visible area. The scrollbar uses a track character (`|`) with a thumb indicator (`┃`) showing the current viewport position. Scrollbars auto-hide when the content fits within the visible area.

### Resizable Pane Dividers

Pane proportions are adjustable via keyboard:

- **Alt+Left / Alt+Right**: adjusts the horizontal split between panels (5% per press, clamped to 15-85%)
- **Alt+Up / Alt+Down**: adjusts the vertical split between the file list and preview pane (5% per press, clamped to 15-85%)

### Themes

Twin Commander ships with 6 built-in color themes:

| Theme | Description |
|-------|-------------|
| Default | Traditional MC-style blue/cyan |
| Dark | Dark neutral tones |
| Light | Light background |
| Solarized | Ethan Schoonover's precision colors |
| Monokai | Classic dark editor theme |
| Nord | Arctic, north-bluish color palette |

Select a theme via Options > Theme in the menu bar. The selected theme is persisted to the configuration file and restored on next launch.

### Hidden Files

Dotfiles (files starting with `.`) are hidden by default. Press `.` to toggle visibility. Each panel has its own independent toggle. When hidden files are visible, the status bar shows `[H]`.

### Filtering

Press `/` to activate filter mode. Type to narrow the file list in real-time (case-insensitive substring match). Press Enter to keep the filter and return to normal mode. Press Escape to clear the filter and return to normal mode. The `..` entry is never filtered out.

## Examples

### Example 1: Browsing a Project

```
$ cd ~/projects/myapp
$ ./twin-commander
```

Both panels show `~/projects/myapp`. Press Down to move to `src/`, press Enter to navigate in. Press Tab to switch to the right panel. Navigate to `docs/` there. Now you have source code on the left and documentation on the right.

### Example 2: Previewing a File

Select a `.go` file and press Enter. An inline preview appears with syntax highlighting -- keywords in blue, comments in gray, strings in green. Press Enter again to open the fullscreen viewer for a closer look. Press Escape to return to the inline preview, and Escape again to close it.

### Example 3: Finding Files with Filter

Press `/` to enter filter mode. Type `test` to show only files containing "test" in their name. Press Enter to keep the filter active while you browse. The filter clears automatically when you navigate into a subdirectory.

### Example 4: Changing the Theme

Press Alt+O to open the Options menu, then select Theme. Choose from Classic, Dark, Nord, Solarized, or Gruvbox. The theme applies immediately and persists across sessions.

### Example 5: Viewing Hidden Files

Press `.` to reveal dotfiles like `.git/`, `.gitignore`, `.env`. The status bar changes from `4 items, 12.3K` to `[H] 7 items, 16.5K`. Press `.` again to hide them.

### Example 6: Resizing Panes

With the preview open, press Alt+Right to give more horizontal space to the right panel. Press Alt+Down to shrink the file list and expand the preview area. Each press adjusts by 5%.

### Bookmarks

Press `Ctrl+B` to open the bookmarks dialog. From there you can:
- Select a bookmark to jump to it
- Press `a` to add the current directory
- Press `x` to remove the selected bookmark
- Press `1-9` in normal mode to jump directly to a bookmark by number

Bookmarks are persisted in your configuration file.

### Configuration

Settings are stored in `~/.config/twin-commander/config.json`. You can edit them via the Options > Configuration menu (Alt+O) or by editing the file directly:

```json
{
  "theme": "default",
  "show_hidden": false,
  "preview_on_start": false,
  "confirm_delete": true,
  "use_trash": true,
  "default_sort_mode": "name",
  "default_sort_asc": true,
  "default_view_mode": "hybrid",
  "editor_command": "",
  "bookmarks": ["/home/user/projects", "/etc"],
  "nerd_font_dismissed": false
}
```

### Git Integration

When inside a git repository, Twin Commander shows:
- **Branch name** in the status bar
- **File status colors**: modified (orange), added (green), deleted (red), renamed (yellow), untracked (gray)
- **Directory status**: aggregated from all files within
- **Git diff** (Ctrl+G): view the diff for the selected file in the fullscreen viewer
- **Stage/unstage** (gs): toggle git staging for the selected file

## File Display

- Directories: blue, bold, with `/` suffix
- Executable files: green
- Symbolic links: purple
- Broken symlinks / inaccessible: dark gray, `---` for size and date
- Selected row: reverse video
- Active panel border: cyan/aqua
- Inactive panel border: default

Sizes are displayed in human-readable format: `450` (bytes), `4.0K`, `1.3M`, `2.1G`.

Dates use `YYYY-MM-DD` format.

## Sort Order

Entries are sorted: `..` first, then directories (alphabetical, case-insensitive), then files (alphabetical, case-insensitive). Symlinks to directories sort with directories; broken symlinks sort with files.

## Error Messages

| Error | Cause | Resolution |
|-------|-------|------------|
| `Permission denied: <path>` | Pressed Enter on a directory without read permission | Navigate elsewhere or fix permissions |
| `Cannot read directory: <path>` | Directory cannot be read (deleted, permissions) | Refresh with `r` or navigate elsewhere |

## Source Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point |
| `app.go` | Application controller, key handling, layout |
| `panel.go` | File panel model and rendering |
| `entry.go` | File entry representation and directory reading |
| `sort.go` | Multi-mode sorting (name, size, date, extension) |
| `filter.go` | Real-time case-insensitive filter |
| `format.go` | Human-readable size formatting |
| `viewmode.go` | View mode definitions (dual-pane, hybrid tree) |
| `tree.go` | Tree panel for hybrid mode |
| `viewer.go` | Fullscreen file viewer with syntax highlighting |
| `menu.go` | Menu bar with hotkeys and dropdown navigation |
| `dialog.go` | Confirm, error, choice, and input dialogs |
| `fileops.go` | File operations (copy, move, delete, trash, mkdir, rename) |
| `search.go` | Recursive filename search with cancellation |
| `git.go` | Git status detection, diff, staging |
| `bookmark.go` | Directory bookmarks with dialog UI |
| `keys.go` | Multi-key sequence tracking (gg, dd, yy, gs) |
| `config.go` | JSON configuration persistence |
| `theme.go` | Color theme system (6 themes) |
| `scrollbar.go` | Scrollbar wrapper for text views |
| `highlight.go` | Syntax highlighting for 14+ languages |
| `icons.go` | Nerd Font file/directory icons |
| `external.go` | External tool integration (editor, bcomp, clipboard) |
| `util.go` | Binary detection, file head reading |

Test files: `entry_test.go`, `filter_test.go`, `format_test.go`, `panel_test.go` (64 tests).

## Running Tests

```bash
go test -v ./...
```

64 tests covering panel logic, sorting, filtering, formatting, entry reading, and rendering.

## Building

```bash
go build -o twin-commander .
```

Produces a single static binary with no runtime dependencies.
