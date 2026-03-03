# DELIVERABLES - Twin Commander (P015)

## Summary

Twin Commander is a Norton Commander / Midnight Commander style dual-pane terminal file explorer built in Go using the tview library. It provides keyboard-driven navigation, real-time search/filter, syntax-highlighted file preview, fuzzy finder, async directory sizes, workspace tabs, configurable themes, and cross-platform terminal support. Ships as a single static binary with zero runtime dependencies.

## Files Delivered

### Application Source

| File | Purpose | Lines |
|------|---------|-------|
| `main.go` | Entry point | ~14 |
| `app.go` | Application controller, layout, key dispatch, feature integration | ~1785 |
| `keybindings.go` | Key event handling, mode dispatch, shortcut routing | ~662 |
| `panel.go` | Panel state, directory operations, rendering | ~355 |
| `entry.go` | FileEntry struct, ReadEntries, SortEntries | ~101 |
| `tree.go` | Persistent filesystem tree panel with expand/collapse | ~383 |
| `menu.go` | Menu bar with hotkeys and dropdown navigation | ~200 |
| `viewer.go` | Fullscreen file viewer with syntax highlighting | ~300 |
| `dialog.go` | Confirm, error, choice, and input dialogs | ~200 |
| `dialogs.go` | Keybindings dialog, permission dialog, UI helpers | ~438 |
| `fileops.go` | File operations (copy, move, delete, trash+trashinfo, mkdir, rename) | ~400 |
| `filehandlers.go` | File operation handlers with UI integration | ~804 |
| `search.go` | Recursive filename search with cancellation | ~150 |
| `fuzzy.go` | Fuzzy filename matching and search with scoring algorithm | ~174 |
| `dirsize.go` | Async directory size calculation with thread-safe cache | ~131 |
| `workspace.go` | Workspace/tab management with full state save/restore | ~108 |
| `contentgrep.go` | Content search (grep) with binary/size skip, cancel support | ~150 |
| `filter.go` | Advanced filtering (substring, glob, regex, negation, multi-term) | ~100 |
| `selection.go` | Pure multi-file selection model (toggle, visual, invert, pattern) | ~100 |
| `history.go` | Pure browser-style directory history (back/forward) | ~60 |
| `permissions.go` | Permission formatting (`rwxr-xr-x`), octal parsing, chmod | ~80 |
| `commandbar.go` | Shell command parsing, variable expansion, execution | ~80 |
| `sort.go` | Multi-mode sorting (name, size, date, extension) | ~80 |
| `format.go` | Human-readable size formatting | ~22 |
| `viewmode.go` | View mode definitions (dual-pane, hybrid tree) | ~20 |
| `keys.go` | Multi-key sequence tracking (gg, dd, yy, gs) | ~40 |
| `config.go` | JSON configuration persistence | ~100 |
| `theme.go` | Color theme system (6 themes) | ~200 |
| `scrollbar.go` | Scrollbar wrapper for text views | ~50 |
| `highlight.go` | Syntax highlighting for 14+ languages | ~300 |
| `icons.go` | Nerd Font file/directory icons | ~200 |
| `git.go` | Git status detection, diff, staging | ~200 |
| `bookmark.go` | Directory bookmarks with dialog UI | ~150 |
| `external.go` | External tool integration (editor, bcomp, xdg-open, clipboard) | ~100 |
| `util.go` | Binary detection, file head reading | ~40 |

### Test Files

| File | Purpose | Tests |
|------|---------|-------|
| `entry_test.go` | Entry/sort tests | 14 |
| `filter_test.go` | Filter logic tests | 5 |
| `format_test.go` | Size formatting tests | 7 |
| `panel_test.go` | Panel integration tests | 38 |
| `selection_test.go` | Selection model tests | 20 |
| `history_test.go` | History model tests | 8 |
| `permissions_test.go` | Permission formatting tests | 10 |
| `commandbar_test.go` | Command parsing tests | 8 |
| `contentgrep_test.go` | Content search tests | 13 |
| `menu_test.go` | Menu alignment and hotkey tests | 4 |
| `fuzzy_test.go` | Fuzzy matching and search tests | 14 |
| `dirsize_test.go` | Directory size cache tests | 7 |
| `workspace_test.go` | Workspace management tests | 8 |
| `app_integration_test.go` | Integration tests for key sequences | 19 |

**Total: 175 tests, 175 pass, 0 fail**

### Documentation & Config

| File | Purpose |
|------|---------|
| `README.md` | Full user documentation and keyboard reference |
| `docs/README.md` | Documentation copy |
| `docs/DELIVERABLES.md` | This file |
| `docs/lint-output.txt` | Lint/diagnostic output |
| `go.mod` | Module definition |
| `go.sum` | Dependency checksums |
| `dev.sh` | Development run script |

## Technology

- **Language**: Go 1.24
- **TUI Framework**: github.com/rivo/tview v0.42.0
- **Terminal**: github.com/gdamore/tcell/v2 v2.13.8
- **Test runner**: `go test`
- **External dependencies**: tview, tcell (and their transitive deps)

## Test Results (from `go test -v ./...` output)

```
# tests 175
# pass 175
# fail 0
ok  twin-commander  0.148s
```

### Test Coverage by Module

**format_test.go (7 tests)** — Size formatting (bytes, KB, MB/GB, boundary values, zero byte, very large, inaccessible sentinel)

**filter_test.go (5 tests)** — Filter logic (case insensitive, matches nothing, dotdot never filtered, empty query, substring)

**entry_test.go (14 tests)** — Entry struct and sort (directories first, case insensitive, stable sort, broken symlink, basic directory, hidden files, metadata, executable, symlinks, nonexistent dir, date format)

**panel_test.go (38 tests)** — Panel integration (load dir, sort order, root, navigate into/up, cursor position, toggle hidden, filter, status bar, refresh, title updates, inaccessible, rendering, symlinks, empty directory)

**selection_test.go (20 tests)** — Selection model (toggle, visual select, invert, pattern select, clear, count, paths)

**history_test.go (8 tests)** — History model (push, back, forward, clear, boundary conditions)

**permissions_test.go (10 tests)** — Permission formatting (rwx display, octal parsing, special modes)

**commandbar_test.go (8 tests)** — Command parsing (variable expansion, %f/%d/%s substitution, edge cases)

**contentgrep_test.go (13 tests)** — Content search (basic match, binary skip, size skip, cancellation, regex)

**menu_test.go (4 tests)** — Menu alignment (click targets, dropdown offsets, consistency, hotkey mapping)

**fuzzy_test.go (14 tests)** — Fuzzy matching (basic match, no match, case insensitive, empty pattern, contiguous beats scattered, prefix beats middle, word boundary, end-to-end search, cancellation, hidden files, max results, directory results)

**dirsize_test.go (7 tests)** — Directory size cache (basic calculation, nested dirs, cache hit, invalidation, invalidate all, cancel, batch request)

**workspace_test.go (8 tests)** — Workspace management (new has one, add creates, can't remove last, remove adjusts active, remove middle, current, auto name, render tab bar)

**app_integration_test.go (19 tests)** — Integration tests (key sequences gg/G/dd/yy/gs, menu hotkeys, view toggle, search mode, fuzzy mode activation, workspace create, workspace switch preserves path)

## Architecture

```
main.go → NewApp() → App.Run()
                      ├── WorkspaceManager (workspace tabs, state save/restore)
                      ├── MenuBar (6 menus, hotkeys, mouse support)
                      ├── TreePanel (persistent filesystem hierarchy)
                      ├── LeftPanel/RightPanel (tview.Table + Panel state)
                      │   ├── Selection model
                      │   ├── History model
                      │   ├── DirSizeCache (async background dir sizes)
                      │   └── Permissions
                      ├── Preview/Viewer (syntax highlighting, scrollbar)
                      ├── FuzzyFinder (Ctrl+P, debounced async search)
                      ├── Search (Ctrl+F, recursive filename)
                      ├── ContentGrep (Ctrl+/, file content search)
                      ├── CommandBar (shell commands)
                      ├── Bookmarks (1-9, Ctrl+B dialog)
                      └── InputCapture (key dispatch: normal/filter/search/fuzzy modes)

Panel.LoadDir() → ReadEntries() → SortEntries() → FilterEntries() → renderTable()
                                                                      └── DirSizeCache.RequestSizesForDir()
```

## Key Features Added (PRs #7-#10)

### PR #7: Mouse Menu Alignment Fix
- Fixed off-by-one in `DropdownOffset()` and `MenuIndexAtX()` — click targets now align correctly with rendered menu items

### PR #8: Fuzzy Finder (Ctrl+P)
- Pure Go fuzzy scoring algorithm (contiguous, word-boundary, prefix, case bonuses)
- Async filesystem search with 150ms debounce
- Tab toggles input/results, Enter navigates, Esc closes

### PR #9: Directory Size Visualization
- Thread-safe `DirSizeCache` with async background calculation
- Live-updating size column (`<DIR>` → `...` → actual size)
- Cache invalidation on file operations

### PR #10: Workspace Tabs
- Full state save/restore (panel paths, sort, hidden, view mode, tree state, splits)
- Tab bar auto-shows with 2+ workspaces, auto-hides with 1
- Ctrl+N create, Ctrl+W close, Alt+1-9 switch

## Build & Run

```bash
go build -o twin-commander .
go test -v ./...
./twin-commander
```

## Project Status

**Active development.** All 20+ functional requirements implemented. 175 tests pass. Binary builds and runs. Full documentation with keyboard reference.
