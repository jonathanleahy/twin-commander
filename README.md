# Twin Commander

A Norton Commander / Midnight Commander style dual-pane terminal file explorer built in Go. Keyboard-driven navigation with real-time search/filter, syntax-highlighted file preview, configurable themes, and cross-platform terminal support.

## Quick Start

```bash
go build -o twin-commander .
./twin-commander
```

Press `q` to exit.

For a comprehensive walkthrough, see the [User Guide](docs/USER_GUIDE.md).

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

- **Persistent filesystem tree browser** — expand/collapse directories in-place, VS Code-style
- **Hybrid Tree + Dual-Pane modes** — toggle with Ctrl+T
- **Quick jump** — `~` to home (preserves tree state), `\` to root (works in all modes), `Ctrl+L` go to path
- **Directory history** — browser-style back (`-`) and forward (`=`) navigation
- **Inline file preview** with syntax highlighting for 14+ languages
- **Fullscreen viewer** with progressive disclosure (preview → viewer → close)
- **Menu bar** with keyboard hotkeys (Alt+F/V/S/G/T/O) and vim-style navigation
- **Multi-file selection** — `Space` to toggle, `v` for visual select, `*` to invert, `+` to select by pattern
- **File operations** — copy, move, rename, delete (with trash + .trashinfo), mkdir — all multi-select aware
- **Cross-filesystem move** — automatic copy+delete fallback when `os.Rename` fails across devices
- **Symlink-safe copy** — `CopyDir` preserves symlinks instead of following them into infinite loops
- **Path traversal protection** — mkdir and rename reject `../../` escape attempts
- **Recursive filename search** with real-time results (Ctrl+F)
- **Fuzzy finder** — Ctrl+P fuzzy filename search with smart scoring (contiguous, word-boundary, prefix bonuses)
- **Directory jump** — `gd` fuzzy directory finder to quickly jump to any directory
- **Content search (grep)** — search file contents with Ctrl+/, skips binary and large files
- **Directory size visualization** — async background calculation with caching, live-updating size column
- **Advanced filtering** — glob patterns (`*.go`), regex (`/pattern/`), negation (`!*.tmp`), multi-term
- **Shell command bar** — run commands with `:`, use `%f` (file), `%d` (dir), `%s` (selected files)
- **Permissions column** — `rwxr-xr-x` display in file list, chmod dialog via menu
- **Git integration** — branch display, file status colors, diff viewer, stage/unstage
- **Directory bookmarks** — save, jump (1-9 keys), manage via Ctrl+B dialog
- **6 color themes** — Default, Dark, Light, Solarized, Monokai, Nord
- **Vim-style keys** — j/k/h/l navigation, gg/G jumps, dd delete, yy yank, p paste
- **Resizable panes** — Alt+arrows to adjust splits (15-85% range)
- **Nerd Font icons** for 60+ file types and directories
- **Beyond Compare integration** — compare files across panels with `b`
- **$EDITOR integration** — open files in your editor with `e` (respects config `editor_command`)
- **xdg-open / open** — open files with system default app with `o`
- **Clipboard support** — copy paths with Alt+C / Opt+C (xclip/xsel/wl-copy/pbcopy)
- **macOS-friendly** — menu hotkeys show Opt instead of Alt on macOS
- **Anchor (scope lock)** — press `a` to lock searches and navigation to a subtree, press `a` again to release
- **Workspace tabs** — Ctrl+N creates workspaces, Alt+1-9 switches, each saves full panel state
- **Session persistence** — workspaces, paths, anchor state, and view settings restored across sessions
- **Configurable** — persistent JSON config for all preferences
- **FreeDesktop trash compliance** — creates `.trashinfo` files for desktop manager restore support

## Usage

Launch Twin Commander from any directory. Optionally pass a starting directory as a CLI argument:

```bash
./twin-commander              # starts in current directory
./twin-commander ~/projects   # starts in ~/projects
```

Session state (workspaces, paths, view mode, anchor) is automatically saved to `~/.config/twin-commander/session.json` and restored on next launch.

The default view is **hybrid tree mode**: a persistent directory tree on the left (rooted at `$HOME`) and a file panel on the right. The tree auto-expands to show your current working directory. Press `Ctrl+T` to switch to classic dual-pane mode.

> **macOS note**: All `Alt+key` shortcuts use `Opt+key` on macOS. Menu labels update automatically.

### Keyboard Reference

#### Navigation

| Key | Action |
|-----|--------|
| j / Down | Move cursor down |
| k / Up | Move cursor up |
| h / Backspace | Collapse tree node / navigate to parent |
| l / Enter | Navigate into directory / expand node / open file |
| gg | Jump to top (works in tree mode too) |
| G | Jump to bottom |
| ~ | Jump to $HOME (preserves tree expanded state) |
| \ | Jump to / (works in both tree and dual-pane mode) |
| a | Anchor — lock scope to current directory |
| - | History back |
| = | History forward |
| Ctrl+L | Go to path (input dialog) |
| gd | Directory jump (fuzzy finder for directories) |
| gr | Recent directories |
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

#### Workspaces

| Key | Action |
|-----|--------|
| Ctrl+N | Create new workspace |
| Ctrl+W | Close current workspace |
| Alt+1-9 (Opt on Mac) | Switch to workspace by number |

#### Selection

| Key | Action |
|-----|--------|
| Space | Toggle select current file + move cursor down |
| v | Start visual (range) selection |
| V / Esc | End visual selection (keeps selection) |
| * | Invert selection |
| + | Select by pattern (substring match dialog) |

#### Search & Filter

| Key | Action |
|-----|--------|
| / | Filter mode (supports glob `*.go`, regex `/pattern/`, negation `!*.tmp`, multi-term) |
| Ctrl+F / F3 | Recursive filename search |
| Ctrl+/ | Content search (grep through file contents) |
| Ctrl+P | Fuzzy finder (smart filename search) |

#### File Operations

| Key | Action |
|-----|--------|
| F5 / c | Copy to other pane (multi-select aware) |
| F6 / m | Move to other pane (multi-select aware) |
| F7 / n | Create new directory (path traversal protected) |
| N | Create new empty file |
| F8 / dd | Delete — trash or permanent (multi-select aware) |
| F2 / R | Rename (path traversal protected) |
| yy | Yank — mark for copy (multi-select aware) |
| p | Paste yanked files |

#### Tools

| Key | Action |
|-----|--------|
| e | Open in $EDITOR (respects config `editor_command`) |
| o | Open with system default (xdg-open / open) |
| : | Run shell command (`%f`=file, `%d`=dir, `%s`=selected) |
| b | Launch Beyond Compare |
| Ctrl+G | Git diff for selected file |
| gs | Git stage/unstage |
| Alt+C (Opt+C on Mac) | Copy path to clipboard |

#### Resize

| Key | Action |
|-----|--------|
| Alt+Left/Right (Opt on Mac) | Adjust horizontal split (5% per press) |
| Alt+Up/Down (Opt on Mac) | Adjust vertical split (file list vs preview) |

#### Menu & System

| Key | Action |
|-----|--------|
| Alt+F/V/S/G/T/O (Opt on Mac) | Open menu by hotkey |
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

### Tree Browsing

The tree panel is a persistent hierarchy rooted at `$HOME` by default. It does not reset on every navigation:

- **Enter on a directory**: expands it in-place (or collapses if already expanded)
- **Enter on a file**: opens an inline preview (or escalates to fullscreen viewer if preview is already open)
- **h / Backspace**: if the node is expanded, collapses it; otherwise moves the cursor to its parent
- **~**: jumps the tree to `$HOME`
- **\\**: re-roots the tree at `/` for full filesystem browsing
- **Ctrl+L**: "Go to Path" dialog for direct path entry (supports `~` expansion)
- **Bookmarks** (1-9, Ctrl+B): expand the tree to the bookmarked path rather than resetting the root

This means you can have `/home/user/projects/` expanded while simultaneously browsing `/home/user/documents/` — the tree preserves all expanded state.

### Navigation (Dual-Pane Mode)

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

Press `/` to activate filter mode. The filter supports multiple matching modes:

- **Substring** (default): `test` matches any file containing "test" (case-insensitive)
- **Glob**: `*.go` matches files ending in `.go` (uses `filepath.Match`)
- **Regex**: `/^test.*\.go$/` matches using Go regular expressions (wrap in `/`)
- **Negation**: `!*.tmp` excludes matching files
- **Multi-term**: `go txt` matches files containing "go" OR "txt"

Press Enter to keep the filter and return to normal mode. Press Escape to clear the filter and return to normal mode. The `..` entry is never filtered out.

## Examples

### Example 1: Browsing a Project

```
$ cd ~/projects/myapp
$ ./twin-commander
```

The tree shows your home directory with `~/projects/myapp` pre-expanded. Press Enter on `src/` to expand it — it stays expanded while you browse other directories. Press Tab to switch to the file panel on the right. Navigate to `docs/` there. The tree preserves all expanded directories, giving you a full project overview.

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

### Fuzzy Finder

Press `Ctrl+P` to open the fuzzy finder. Type any part of a filename to search — the fuzzy matching algorithm scores results based on:

- **Contiguous character matches** (escalating bonus for consecutive matches)
- **Word boundary matches** (bonus for matches after `/`, `.`, `_`, `-`)
- **Filename prefix** (bonus for matching at the start of the filename)
- **Exact case** (small bonus for matching case exactly)
- **Path length** (shorter paths rank higher)

Results update as you type (150ms debounce). Press `Tab` to toggle between the input field and results table. Press `Enter` to navigate to the selected result. Press `Esc` to close.

### Directory Sizes

Directory sizes are calculated asynchronously in the background. When you navigate to a directory:
- Directories initially show `<DIR>` in the size column
- As sizes are calculated, the column updates to show `...` (calculating) then the actual size (e.g., `4.2M`)
- Sizes are cached and reused until a file operation invalidates the cache
- The size includes all files recursively within the directory

### Workspaces

Workspaces let you maintain multiple independent browsing sessions. Each workspace saves:
- Both panels' paths, sort mode, hidden file settings
- View mode (tree/dual), active panel, preview state
- Tree root path and expanded directories
- Pane split proportions

| Action | Key |
|--------|-----|
| Create new workspace | `Ctrl+N` |
| Close current workspace | `Ctrl+W` |
| Switch to workspace 1-9 | `Alt+1` through `Alt+9` |

The tab bar appears between the menu bar and content area when you have more than one workspace. It auto-hides when only one workspace remains.

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
  "tree_root": "home",
  "bookmarks": ["/home/user/projects", "/etc"],
  "nerd_font_dismissed": false
}
```

- `tree_root`: set to `"/"` to start the tree at the filesystem root instead of `$HOME`

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
| `app.go` | Application controller, key handling, layout, feature integration |
| `panel.go` | File panel model, rendering, selection/history/permissions integration |
| `entry.go` | File entry representation with Mode field and directory reading |
| `selection.go` | Pure multi-file selection model (toggle, visual, invert, pattern) |
| `history.go` | Pure browser-style directory history (back/forward) |
| `permissions.go` | Permission formatting (`rwxr-xr-x`), octal parsing, chmod |
| `commandbar.go` | Shell command parsing, variable expansion, execution |
| `contentgrep.go` | Content search (grep) with binary/size skip, cancel support |
| `sort.go` | Multi-mode sorting (name, size, date, extension) |
| `filter.go` | Advanced filtering (substring, glob, regex, negation, multi-term) |
| `format.go` | Human-readable size formatting |
| `viewmode.go` | View mode definitions (dual-pane, hybrid tree) |
| `tree.go` | Persistent filesystem tree panel with expand/collapse |
| `viewer.go` | Fullscreen file viewer with syntax highlighting |
| `menu.go` | Menu bar with hotkeys and dropdown navigation |
| `dialog.go` | Confirm, error, choice, and input dialogs |
| `fileops.go` | File operations (copy, move, delete, trash+trashinfo, mkdir, rename) |
| `search.go` | Recursive filename search with cancellation |
| `git.go` | Git status detection, diff, staging |
| `bookmark.go` | Directory bookmarks with dialog UI |
| `keys.go` | Multi-key sequence tracking (gg, dd, yy, gs) |
| `config.go` | JSON configuration persistence |
| `theme.go` | Color theme system (6 themes) |
| `scrollbar.go` | Scrollbar wrapper for text views |
| `highlight.go` | Syntax highlighting for 14+ languages |
| `icons.go` | Nerd Font file/directory icons |
| `fuzzy.go` | Fuzzy filename matching and search with scoring algorithm |
| `dirsize.go` | Async directory size calculation with thread-safe cache |
| `workspace.go` | Workspace/tab management with full state save/restore |
| `session.go` | Session persistence — save/restore workspaces across launches |
| `filehandlers.go` | File operation handlers (copy, move, delete, rename, mkdir) |
| `external.go` | External tool integration (editor, bcomp, xdg-open, clipboard) |
| `util.go` | Binary detection, file head reading |

Test files: `entry_test.go`, `filter_test.go`, `format_test.go`, `panel_test.go`, `selection_test.go`, `history_test.go`, `permissions_test.go`, `commandbar_test.go`, `contentgrep_test.go`, `menu_test.go`, `fuzzy_test.go`, `dirsize_test.go`, `workspace_test.go`, `session_test.go`, `app_integration_test.go` (191 tests).

## Running Tests

```bash
go test -v ./...
```

191 tests covering selection model, history model, permissions, command parsing, content search, panel logic, sorting, filtering, formatting, entry reading, rendering, menu alignment, fuzzy matching, directory size calculation, workspace management, session persistence, anchor scope lock, and integration tests for key sequences.

## Building

```bash
go build -o twin-commander .
```

Produces a single static binary with no runtime dependencies.
