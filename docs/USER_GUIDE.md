# Twin Commander User Guide

A comprehensive guide to Twin Commander, a keyboard-driven dual-pane terminal file manager inspired by Norton Commander and Midnight Commander.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Interface Overview](#interface-overview)
3. [Navigation](#navigation)
4. [View Modes](#view-modes)
5. [File Preview & Viewer](#file-preview--viewer)
6. [File Selection](#file-selection)
7. [File Operations](#file-operations)
8. [Searching & Filtering](#searching--filtering)
9. [Fuzzy Finder](#fuzzy-finder)
10. [Directory Jump](#directory-jump)
11. [Directory Sizes](#directory-sizes)
12. [Anchor (Scope Lock)](#anchor-scope-lock)
13. [Workspaces](#workspaces)
14. [Bookmarks](#bookmarks)
15. [Shell Command Bar](#shell-command-bar)
16. [Git Integration](#git-integration)
17. [Themes & Appearance](#themes--appearance)
18. [Configuration](#configuration)
19. [Menu Bar Reference](#menu-bar-reference)
20. [Complete Keyboard Reference](#complete-keyboard-reference)
21. [macOS Notes](#macos-notes)
22. [Troubleshooting](#troubleshooting)

---

## Getting Started

### Installation

**Prerequisites**: Go 1.22 or later.

```bash
# Build from source
go build -o twin-commander .

# Run
./twin-commander

# Exit
# Press q
```

Twin Commander launches as a full-screen terminal application. You can optionally pass a starting directory:

```bash
./twin-commander              # starts in current directory
./twin-commander ~/projects   # starts in ~/projects
```

It starts in hybrid tree mode showing a directory tree rooted at `$HOME` on the left and a file panel on the right, pre-expanded to your current working directory. Session state is automatically saved and restored across launches.

### First Steps

1. **Move around** with `j`/`k` (or arrow keys) to move the cursor
2. **Enter a directory** with `l` or `Enter`
3. **Go back** with `h` or `Backspace`
4. **Switch panes** with `Tab`
5. **Preview a file** by pressing `Enter` on any text file
6. **Quit** with `q`

---

## Interface Overview

### Layout

```
┌─────────────────────────────────────────────────────────┐
│  File  View  Search  Go  Tools  Options                 │  ← Menu bar
├───────────────────┬─────────────────────────────────────┤
│ ~/                │ Name           Size    Date         │
│ ├── .config/      │ ..                                  │
│ ├── Documents/    │ src/           <DIR>   2025-01-15   │
│ │   ├── notes/    │ README.md      4.2K   2025-01-14   │
│ │   └── photos/   │ main.go        1.3K   2025-01-13   │
│ ├── projects/     │ go.mod           89   2025-01-10   │
│ │   └── myapp/    │                                     │
│ └── work/         │                                     │
│                   │                                     │
├───────────────────┴─────────────────────────────────────┤
│ ~/projects/myapp                   6 items, 12.4K       │  ← Status bar
└─────────────────────────────────────────────────────────┘
```

### Components

- **Menu bar** (top) — Six menus accessible via Alt+hotkey or F9/F10. Click with mouse to open.
- **Tree panel** (left, hybrid mode) — Persistent directory hierarchy. Expand/collapse directories in-place.
- **File panel** (right, or both sides in dual mode) — File listing with Name, Size, Date columns. Permissions column visible too.
- **Status bar** (bottom) — Shows current path, item count, total size, sort mode, and selection count. Shows `[H]` when hidden files are visible.
- **Tab bar** (between menu and content) — Appears when you have 2+ workspaces. Shows workspace numbers with the active one highlighted.

### File Display Colors

| Color | Meaning |
|-------|---------|
| Blue, bold | Directory |
| Green | Executable file |
| Purple | Symbolic link |
| Dark gray | Broken symlink or inaccessible |
| Orange | Git modified |
| Green (git) | Git added/staged |
| Red (git) | Git deleted |
| Yellow (git) | Git renamed |
| Gray (git) | Git untracked |

### Size & Date Format

- Sizes: `450` (bytes), `4.0K`, `1.3M`, `2.1G` — human-readable
- Dates: `YYYY-MM-DD` format
- Directories: show `<DIR>`, then async-calculated total size (e.g., `4.2M`)

### Nerd Font Icons

Twin Commander displays icons for 60+ file types and special directories when a Nerd Font is installed. On first launch without a Nerd Font, a one-time dialog offers installation guidance. Special folder icons include `.git`, `node_modules`, `src`, `test`, `docs`, `build`, `.github`, `.vscode`, and `config`.

---

## Navigation

### Basic Movement

| Key | Action |
|-----|--------|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `h` / `Backspace` | Go to parent directory / collapse tree node |
| `l` / `Enter` | Enter directory / expand tree node / open file preview |
| `gg` | Jump to first item |
| `G` | Jump to last item |

### Quick Jump

| Key | Action |
|-----|--------|
| `~` | Jump to `$HOME` (preserves all tree expanded state) |
| `\` | Jump to `/` root filesystem (works in both modes) |
| `Ctrl+L` | Go to path — type any absolute path or use `~` for home |
| `gd` | Directory jump — fuzzy finder for directories only |
| `gr` | Recent directories — list of recently visited dirs |
| `a` | Anchor — lock scope to current directory |
| `-` | History back (like a browser back button) |
| `=` | History forward |

### Pane Switching

| Key | Action |
|-----|--------|
| `Tab` | Cycle to next pane (tree → files → preview → tree) |
| `Shift+Tab` | Cycle to previous pane |

In **hybrid mode**, Tab cycles: tree → file panel → preview (if open) → tree.
In **dual-pane mode**, Tab cycles: left panel → right panel → preview (if open) → left panel.

### Directory History

Twin Commander maintains a browser-style history stack per panel. Each time you enter a directory, it's pushed onto the stack. Press `-` to go back and `=` to go forward. Duplicate consecutive entries are suppressed. The history has a max depth to prevent unbounded growth.

---

## View Modes

### Hybrid Tree Mode (Default)

Press `Ctrl+T` to toggle between modes.

The tree panel on the left is a persistent hierarchy rooted at `$HOME`. Unlike traditional file managers, it does **not** reset when you navigate — expanded directories stay expanded:

- **Enter on a directory** expands it in-place (or collapses if already expanded)
- **Enter on a file** opens an inline preview
- **h / Backspace** collapses an expanded node, or moves to parent if already collapsed
- **~** jumps to `$HOME` while preserving all expanded directories
- **\\** re-roots the tree at `/` for full filesystem browsing

This means you can have `/home/user/projects/` expanded while simultaneously browsing `/home/user/documents/`.

### Dual-Pane Mode

Classic two-panel layout. Both panels show full file listings side by side. Each panel navigates independently. File operations (copy, move) work between the two panels.

### Resizing Panes

| Key | Action |
|-----|--------|
| `Alt+Left` / `Alt+Right` | Adjust horizontal split (5% per press, 15-85% range) |
| `Alt+Up` / `Alt+Down` | Adjust vertical split between file list and preview |

### Sort Options

| Key | Action |
|-----|--------|
| `s` | Cycle sort mode: name → size → date → extension → name |
| `S` | Toggle sort order: ascending / descending |

The current sort mode and direction are shown in the status bar (e.g., `name↑`, `size↓`).

Sort always places `..` first, then directories (alphabetical, case-insensitive), then files.

### Hidden Files

Press `.` to toggle hidden file visibility. Each panel has its own independent toggle. When hidden files are visible, the status bar shows `[H]`. This setting can be made default via configuration.

---

## File Preview & Viewer

Twin Commander uses a progressive disclosure model for file viewing:

### Step 1: Inline Preview

Press `Enter` on any text file. A preview pane appears alongside the file list with syntax highlighting. The preview pane is scrollable when focused.

### Step 2: Fullscreen Viewer

Press `Enter` again while the preview is showing. The file opens in a fullscreen viewer with the same syntax highlighting and scrollbar support.

### Step 3: Returning

- `Escape` from fullscreen → returns to inline preview
- `t` from fullscreen → returns to inline preview (keeping preview open)
- `Escape` from inline preview → closes preview entirely

### Preview/Viewer Controls (when focused)

| Key | Action |
|-----|--------|
| `j` / `Down` | Scroll down one line |
| `k` / `Up` | Scroll up one line |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `PgUp` | Page up (20 lines) |
| `PgDn` | Page down (20 lines) |
| `Esc` | Close / unfocus |

### Syntax Highlighting

Highlighting is applied in both inline preview and fullscreen viewer for these file types:

| Language | Extensions |
|----------|-----------|
| Go | `.go` |
| JavaScript/TypeScript | `.js`, `.ts`, `.jsx`, `.tsx`, `.mjs` |
| Python | `.py` |
| Rust | `.rs` |
| C/C++/Java/Kotlin | `.c`, `.h`, `.cpp`, `.hpp`, `.cc`, `.cxx`, `.cs`, `.java`, `.kt` |
| Shell | `.sh`, `.bash`, `.zsh`, `.fish`, `Makefile`, `Dockerfile` |
| YAML/TOML | `.yaml`, `.yml`, `.toml` |
| JSON | `.json` |
| HTML/XML | `.html`, `.htm`, `.xml`, `.svg` |
| CSS/SCSS | `.css`, `.scss`, `.sass`, `.less` |
| SQL | `.sql` |
| Markdown | `.md`, `.markdown` |
| Diff/Patch | `.diff`, `.patch` |

### Scrollbars

Preview panes and the fullscreen viewer display scrollbars when content exceeds the visible area. The scrollbar uses a track (`|`) with a thumb indicator (`┃`) showing viewport position. Scrollbars auto-hide when content fits.

---

## File Selection

### Single Selection

Move the cursor with `j`/`k` or arrow keys. The highlighted row is the current selection. Most file operations act on the current file when no multi-selection is active.

### Multi-Selection

| Key | Action |
|-----|--------|
| `Space` | Toggle select current file and move cursor down |
| `v` | Start visual (range) selection from current position |
| `V` or `Esc` | End visual selection (keeps files selected) |
| `*` | Invert selection — selects all unselected, deselects all selected |
| `+` | Select by pattern — opens a dialog for substring matching |

When files are selected, the status bar shows the selection count (e.g., `3 selected`). File operations (copy, move, delete, yank) automatically operate on all selected files.

### Visual Selection

1. Move cursor to the start of your range
2. Press `v` to begin visual selection
3. Move cursor to the end of your range — files are highlighted as you move
4. Press `V` or `Esc` to finalize (all files in the range stay selected)

---

## File Operations

All file operations are multi-select aware. If files are selected, the operation applies to all selected files. Otherwise, it applies to the file under the cursor.

| Key | Action |
|-----|--------|
| `F5` / `c` | **Copy** to other pane |
| `F6` / `m` | **Move** to other pane |
| `F7` / `n` | **Create** new directory |
| `N` | **Create** new empty file |
| `F8` / `dd` | **Delete** (trash or permanent, configurable) |
| `F2` / `R` | **Rename** current file/directory |
| `yy` | **Yank** — mark for copy (vim-style) |
| `p` | **Paste** yanked files |

### Copy & Move

Files are copied/moved to the directory shown in the other panel (or the opposite panel in dual-pane mode). A confirmation dialog appears if a file would be overwritten. Cross-filesystem moves are handled automatically — if `os.Rename` fails across devices, Twin Commander falls back to copy + delete.

### Delete

By default, Twin Commander uses **soft delete** (trash). Deleted files go to `~/.local/share/Trash/files/` with a `.trashinfo` file created for desktop manager restore support (FreeDesktop compliant). You can switch to permanent delete in configuration or per-operation.

The `confirm_delete` config option controls whether a confirmation dialog appears.

### Rename

Opens an input dialog pre-filled with the current name. Path traversal attempts (`../../`) are rejected for security.

### Create Directory

Opens an input dialog for the new directory name. Created in the current panel's directory. Path traversal attempts are rejected.

### Yank & Paste (Vim-style)

Press `yy` to mark the current file (or all selected files) for copying. Navigate to the destination directory and press `p` to paste. This is an internal clipboard — it doesn't use the system clipboard.

### Symlink Safety

`CopyDir` preserves symlinks instead of following them, preventing infinite loops when copying directory trees that contain circular symlinks.

---

## Searching & Filtering

### Filter Mode (`/`)

Press `/` to enter filter mode. Type to narrow the file list in real-time. The filter supports multiple matching modes:

| Pattern | Mode | Example |
|---------|------|---------|
| `test` | Substring (case-insensitive) | Matches `test_utils.go`, `MyTest.java` |
| `*.go` | Glob pattern | Matches all `.go` files |
| `/^test.*\.go$/` | Regex (wrap in `/`) | Matches `test_main.go` but not `my_test.go` |
| `!*.tmp` | Negation | Hides all `.tmp` files |
| `go txt` | Multi-term (space-separated) | Matches files containing "go" OR "txt" |

- Press `Enter` to keep the filter and return to normal mode
- Press `Escape` to clear the filter and return to normal mode
- The `..` entry is never filtered out
- Filters clear automatically when navigating into a subdirectory

### Recursive Filename Search (`Ctrl+F` / `F3`)

Opens a full-screen search overlay. Type a filename pattern and results appear as matching files are found across all subdirectories. Press `Tab` to switch between the search input and results table. Press `Enter` on a result to navigate to that file's directory. Press `Escape` to close.

### Content Search / Grep (`Ctrl+/`)

Searches inside file contents. Results show matching files with line numbers and context. Binary files and files larger than 1MB are automatically skipped. Supports cancellation — press `Escape` to stop a running search.

---

## Fuzzy Finder

Press `Ctrl+P` to open the fuzzy finder — a fast, VS Code-style file search.

### How It Works

Type any part of a filename. The fuzzy matching algorithm finds files where all your typed characters appear in order (not necessarily consecutive) and ranks them by quality:

- **Contiguous matches** score highest (escalating bonus for consecutive characters)
- **Word boundary matches** get a bonus (after `/`, `.`, `_`, `-`)
- **Filename prefix matches** get a bonus (matching the start of the filename)
- **Exact case matches** get a small bonus
- **Shorter paths** rank higher than deeper ones

Results update as you type with a 150ms debounce to keep the UI responsive.

### Controls

| Key | Action |
|-----|--------|
| Type | Filter results in real-time |
| `Tab` | Toggle focus between input field and results table |
| `j` / `k` | Navigate results (when table is focused) |
| `Enter` | Navigate to selected file |
| `Esc` | Close fuzzy finder |

### Example

Typing `apgo` would match `app.go` (contiguous match), `app_config.go` (word boundary), and `application/handler.go` (scattered match) — ranked in that order.

---

## Directory Jump

Press `gd` to open the directory jump overlay — a fuzzy finder that shows **only directories**. This is useful when you know part of a directory name and want to jump to it quickly without navigating through the tree manually.

### Controls

| Key | Action |
|-----|--------|
| Type | Filter directories in real-time |
| `Tab` | Toggle focus between input field and results table |
| `j` / `k` | Navigate results (when table is focused) |
| `Enter` | Jump to selected directory |
| `Esc` | Close directory jump |

When anchored, results are scoped to the anchor directory.

---

## Directory Sizes

Directory sizes are calculated asynchronously in the background. This means the UI stays responsive while sizes are computed.

### How It Works

1. When you navigate to a directory, subdirectories initially show `<DIR>` in the size column
2. Background goroutines walk each subdirectory, summing all file sizes recursively
3. While calculating, the size column shows `...`
4. Once complete, the actual size appears (e.g., `4.2M`)
5. Sizes are cached — revisiting a directory shows sizes instantly
6. The cache is invalidated automatically after file operations (copy, move, delete)

### Sorting by Size

When you sort by size (`s` to cycle to size mode), directories are sorted by their calculated sizes. Directories whose sizes haven't been calculated yet sort as zero.

---

## Anchor (Scope Lock)

The Anchor feature lets you lock your working scope to a specific directory. When anchored, all searches, fuzzy finder results, and navigation are constrained to the anchored subtree.

### Usage

1. Navigate to the directory you want to scope
2. Press `a` to anchor — the status bar shows `⚓` and the anchor path
3. Press `a` again to release the anchor

### What Gets Scoped

When anchored:

- **Recursive search** (`Ctrl+F`): only searches within the anchor directory
- **Fuzzy finder** (`Ctrl+P`): only indexes files within the anchor directory
- **Directory jump** (`gd`): only indexes directories within the anchor directory
- **Navigate up** (`h` / `Backspace`): stops at the anchor root — you can't go higher
- **Jump to home** (`~`): jumps to the anchor root instead of `$HOME`
- **Jump to root** (`\`): jumps to the anchor root instead of `/`
- **Go to path** (`Ctrl+L`): blocks paths outside the anchor scope
- **Bookmarks** (`1-9`): blocks bookmarks that point outside the anchor scope
- **Tree view**: re-roots to the anchor directory in hybrid mode

### Workspace Persistence

Anchor state is saved per workspace. Switching workspaces preserves each workspace's anchor independently.

---

## Workspaces

Workspaces let you maintain multiple independent browsing sessions within a single Twin Commander instance. Each workspace saves its complete state independently.

### Creating & Switching

| Key | Action |
|-----|--------|
| `Ctrl+N` | Create a new workspace |
| `Ctrl+W` | Close current workspace (cannot close the last one) |
| `Alt+1` through `Alt+9` | Switch to workspace by number |

### What Each Workspace Saves

- Both panels' directory paths
- Sort mode and sort order per panel
- Hidden file visibility per panel
- View mode (hybrid tree or dual-pane)
- Active panel (left or right)
- Preview pane state
- Tree root path and all expanded directories
- Horizontal and vertical split proportions
- Anchor path and active state

### Session Persistence

All workspace state is automatically saved to `~/.config/twin-commander/session.json` when you exit and restored on next launch. This includes all workspaces, their paths, view modes, expanded tree nodes, anchor state, and split proportions. Invalid paths (deleted directories) are gracefully handled by falling back to `$HOME`.

### Tab Bar

The tab bar appears between the menu bar and content area when you have 2 or more workspaces. The active workspace is highlighted. When you close workspaces down to one, the tab bar auto-hides to save screen space.

### Typical Workflow

1. Start browsing a project in workspace 1
2. Press `Ctrl+N` to create workspace 2 for a different task
3. Navigate to a different directory in workspace 2
4. Press `Alt+1` to switch back to workspace 1 — everything is exactly where you left it
5. Press `Alt+2` to switch back to workspace 2
6. Press `Ctrl+W` to close workspace 2 when done

---

## Bookmarks

Save up to 9 directory bookmarks for quick access.

### Using Bookmarks

| Key | Action |
|-----|--------|
| `1` through `9` | Jump to bookmark by number |
| `Ctrl+B` | Open bookmark manager dialog |

### Bookmark Manager

Press `Ctrl+B` to open the manager:

- **Select a bookmark** and press `Enter` to navigate to it
- Press `a` to add the current directory as a new bookmark
- Press `x` to remove the highlighted bookmark
- Press `Escape` to close

In hybrid tree mode, jumping to a bookmark expands the tree to show the bookmarked path rather than resetting the tree root.

Bookmarks are saved in your configuration file and persist across sessions.

---

## Shell Command Bar

Press `:` to open the command bar. Type a shell command and press `Enter` to execute it.

### Variable Substitution

| Variable | Expands To |
|----------|-----------|
| `%f` | Path of the current file |
| `%d` | Current directory path |
| `%s` | All selected file paths (space-separated) |

### Examples

```
:cat %f                    # Display current file contents
:wc -l %s                  # Count lines in all selected files
:find %d -name "*.log"     # Find log files in current directory
:chmod 755 %f              # Change permissions on current file
:git log --oneline %f      # Git log for current file
```

Command output is displayed in the fullscreen viewer. Press `Escape` to return to the file manager after viewing output.

---

## Git Integration

When inside a git repository, Twin Commander provides:

### Status Display

- **Branch name** shown in the status bar
- **File status colors** in the file listing:
  - Orange — modified
  - Green — added/staged
  - Red — deleted
  - Yellow — renamed
  - Gray — untracked
- **Directory status** — aggregated from all contained files

### Git Commands

| Key | Action |
|-----|--------|
| `Ctrl+G` | View git diff for the current file in the fullscreen viewer |
| `gs` | Toggle git stage/unstage for the current file |

---

## Themes & Appearance

Twin Commander ships with 6 color themes.

### Changing Themes

Open the Options menu (`Alt+O`) and select **Theme**, or use the Options > Theme submenu. The theme applies immediately and is saved to your configuration file.

### Available Themes

| Theme | Description |
|-------|-------------|
| Default | Traditional MC-style — cyan borders, navy menu bar |
| Dark | Neutral dark — green borders, black menu bar |
| Light | Light background — blue borders, white menu bar |
| Solarized | Ethan Schoonover's precision colors — blue accent, warm tones |
| Monokai | Classic editor theme — green borders, dark gray bar |
| Nord | Arctic palette — frost blue borders, polar night bar |

---

## Configuration

### Config File Location

```
~/.config/twin-commander/config.json
```

Edit via the Options > Configuration menu (`Alt+O`) or by editing the file directly.

### All Configuration Options

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
  "bookmarks": [],
  "nerd_font_dismissed": false
}
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `theme` | string | `"default"` | Color theme: `default`, `dark`, `light`, `solarized`, `monokai`, `nord` |
| `show_hidden` | bool | `false` | Show dotfiles by default on startup |
| `preview_on_start` | bool | `false` | Open preview pane automatically on startup |
| `confirm_delete` | bool | `true` | Show confirmation dialog before deleting |
| `use_trash` | bool | `true` | Soft delete to trash instead of permanent delete |
| `default_sort_mode` | string | `"name"` | Default sort: `name`, `size`, `date`, `extension` |
| `default_sort_asc` | bool | `true` | Sort ascending by default |
| `default_view_mode` | string | `"hybrid"` | Startup mode: `hybrid` (tree+panel) or `dual` (two panels) |
| `editor_command` | string | `""` | Custom editor command. Falls back to `$EDITOR`, then `vi` |
| `tree_root` | string | `"home"` | Tree panel root: `"home"` for `$HOME`, `"/"` for filesystem root |
| `bookmarks` | array | `[]` | Saved directory paths (up to 9) |
| `nerd_font_dismissed` | bool | `false` | Suppress the Nerd Font installation reminder |

---

## Menu Bar Reference

Open any menu with its Alt+hotkey, or press `F9`/`F10` to activate the menu bar. Navigate menus with arrow keys or `h`/`j`/`k`/`l`. Click with mouse to open menus and select items.

### File (Alt+F)

| Item | Shortcut |
|------|----------|
| New Directory | F7 / n |
| New File | N |
| Copy | F5 / c |
| Move | F6 / m |
| Rename | F2 / R |
| Delete | F8 / dd |
| Quit | q / Ctrl+C |

### View (Alt+V)

| Item | Shortcut |
|------|----------|
| Toggle Tree/Dual | Ctrl+T |
| Toggle Hidden Files | . |
| Toggle Preview Pane | t |
| Cycle Sort Mode | s |
| Toggle Sort Order | S |
| Refresh | r |
| New Workspace | Ctrl+N |
| Close Workspace | Ctrl+W |

### Search (Alt+S)

| Item | Shortcut |
|------|----------|
| Filter (glob/regex) | / |
| Recursive Search | Ctrl+F / F3 |
| Content Search | Ctrl+/ |
| Fuzzy Finder | Ctrl+P |

### Go (Alt+G)

| Item | Shortcut |
|------|----------|
| Anchor | a |
| Go to Path... | Ctrl+L |
| Directory Jump | gd |
| Recent Dirs | gr |
| Jump to Home | ~ |
| Jump to Root / | \ |
| History Back | - |
| History Forward | = |
| Bookmarks... | Ctrl+B |
| Jump to Bookmark 1-3 | 1, 2, 3 |

### Tools (Alt+T)

| Item | Shortcut |
|------|----------|
| Open in Editor | e |
| Open with Default | o |
| View File | Enter |
| Shell Command | : |
| Change Permissions | (menu only) |
| Beyond Compare | b |
| Copy Path | Alt+C |
| Git Diff | Ctrl+G |
| Git Stage/Unstage | gs |

### Options (Alt+O)

| Item | Description |
|------|-------------|
| Theme... | Select from 6 color themes |
| Configuration... | Edit settings (checkboxes, dropdowns, text fields) |
| Key Bindings... | View complete keyboard reference |
| About... | Application info |

---

## Complete Keyboard Reference

### Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `h` / `Backspace` | Collapse tree node / navigate to parent |
| `l` / `Enter` | Enter directory / expand node / open file |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `~` | Jump to $HOME |
| `\` | Jump to / |
| `a` | Anchor (scope lock) |
| `-` | History back |
| `=` | History forward |
| `Ctrl+L` | Go to path |
| `gd` | Directory jump (fuzzy finder for directories) |
| `gr` | Recent directories |
| `Tab` | Cycle pane forward |
| `Shift+Tab` | Cycle pane backward |

### View

| Key | Action |
|-----|--------|
| `Ctrl+T` | Toggle hybrid tree / dual-pane mode |
| `t` | Toggle inline preview pane |
| `.` | Toggle hidden files |
| `s` | Cycle sort mode (name / size / date / extension) |
| `S` | Toggle sort order (ascending / descending) |
| `r` | Refresh current directory |

### Workspaces

| Key | Action |
|-----|--------|
| `Ctrl+N` | Create new workspace |
| `Ctrl+W` | Close current workspace |
| `Alt+1` – `Alt+9` | Switch to workspace by number |

### Selection

| Key | Action |
|-----|--------|
| `Space` | Toggle select + move down |
| `v` | Start visual selection |
| `V` / `Esc` | End visual selection |
| `*` | Invert selection |
| `+` | Select by pattern |

### Search & Filter

| Key | Action |
|-----|--------|
| `/` | Filter mode (glob, regex, negation, multi-term) |
| `Ctrl+F` / `F3` | Recursive filename search |
| `Ctrl+/` | Content search (grep) |
| `Ctrl+P` | Fuzzy finder |

### File Operations

| Key | Action |
|-----|--------|
| `F5` / `c` | Copy to other pane |
| `F6` / `m` | Move to other pane |
| `F7` / `n` | Create new directory |
| `N` | Create new empty file |
| `F8` / `dd` | Delete (trash or permanent) |
| `F2` / `R` | Rename |
| `yy` | Yank (mark for copy) |
| `p` | Paste yanked files |

### Tools

| Key | Action |
|-----|--------|
| `e` | Open in $EDITOR |
| `o` | Open with system default (xdg-open / open) |
| `:` | Shell command bar |
| `b` | Beyond Compare |
| `Ctrl+G` | Git diff |
| `gs` | Git stage/unstage |
| `Alt+C` | Copy path to clipboard |

### Resize

| Key | Action |
|-----|--------|
| `Alt+Left` / `Alt+Right` | Adjust horizontal split (5% per press) |
| `Alt+Up` / `Alt+Down` | Adjust vertical split |

### Menu & System

| Key | Action |
|-----|--------|
| `Alt+F/V/S/G/T/O` | Open menu by hotkey |
| `F9` / `F10` | Open menu bar |
| `Ctrl+B` | Bookmarks dialog |
| `1` – `9` | Jump to bookmark |
| `q` | Quit |
| `Ctrl+C` | Force quit |
| `Esc` | Close overlay / cancel |

### Preview/Viewer Keys

| Key | Action |
|-----|--------|
| `j` / `Down` | Scroll down one line |
| `k` / `Up` | Scroll up one line |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `PgUp` | Page up |
| `PgDn` | Page down |
| `t` | Return from fullscreen to inline preview |
| `Esc` | Close viewer/preview |

### Multi-Key Sequences

These require pressing two keys within 500ms:

| Sequence | Action |
|----------|--------|
| `g` then `g` | Jump to top |
| `d` then `d` | Delete |
| `y` then `y` | Yank (mark for copy) |
| `g` then `s` | Git stage/unstage |

### Menu Navigation Keys

| Key | Action |
|-----|--------|
| `Left` / `h` | Previous menu |
| `Right` / `l` | Next menu |
| `Up` / `k` | Previous item |
| `Down` / `j` | Next item |
| `Enter` | Activate selected item |
| `Esc` / `q` | Close menu |

---

## macOS Notes

### Option Key Mapping

On macOS, the `Option` key sends Unicode characters instead of `Alt` modifier events. Twin Commander maps these automatically:

| macOS Key | Sends | Maps To |
|-----------|-------|---------|
| `Opt+F` | `ƒ` | Alt+F (File menu) |
| `Opt+V` | `√` | Alt+V (View menu) |
| `Opt+S` | `ß` | Alt+S (Search menu) |
| `Opt+G` | `©` | Alt+G (Go menu) |
| `Opt+T` | `†` | Alt+T (Tools menu) |
| `Opt+O` | `ø` | Alt+O (Options menu) |
| `Opt+C` | `ç` | Alt+C (Copy path) |

Menu labels automatically show `Opt` instead of `Alt` on macOS.

### Function Keys

On Mac keyboards with a Touch Bar or compact layout, F-keys may require pressing `Fn` first. Alternatives:
- `F5` → `c` (copy)
- `F6` → `m` (move)
- `F7` → `n` (new directory)
- `F8` → `dd` (delete)
- `F2` → `R` (rename)
- `F9`/`F10` → `Alt+F` or `Opt+F` (menu bar)

### Clipboard

macOS uses `pbcopy` for clipboard operations (`Alt+C`). No additional tools needed.

### Default Opener

`o` key uses the `open` command on macOS (equivalent to `xdg-open` on Linux).

---

## Troubleshooting

### Icons Look Wrong or Missing

Install a Nerd Font. Download from [nerdfonts.com](https://www.nerdfonts.com/) and set it as your terminal's font. On first launch, Twin Commander shows a dialog with platform-specific installation instructions. Check "Don't remind me" to suppress the dialog.

### Alt Keys Don't Work

- **macOS**: Use `Opt` key instead of `Alt`. The mappings are automatic.
- **Linux (some terminals)**: Your terminal may intercept Alt keys. Try `F9` to access the menu bar instead, or check your terminal's key binding settings.
- **tmux**: Alt keys may be captured by tmux. Use `F9`/`F10` for menu access.

### Terminal Too Small

Twin Commander needs a minimum terminal size to render properly. If the layout looks broken, try enlarging your terminal window.

### Files Not Updating

Press `r` to refresh the current directory. File operations from outside Twin Commander (other terminals, scripts) won't be detected automatically.

### Cross-Filesystem Move Fails

Twin Commander automatically falls back to copy+delete when moving files across filesystem boundaries. If this fails, check that you have write permissions on both source and destination.

### Git Status Not Showing

Git integration requires `git` to be installed and accessible in your `$PATH`. The current directory (or a parent) must be inside a git repository.

### Editor Doesn't Open

Editor priority: `editor_command` config → `$EDITOR` environment variable → `vi`. Set your preferred editor in the configuration dialog or `config.json`.
