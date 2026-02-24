# Twin Commander

A Norton Commander / Midnight Commander style dual-pane terminal file explorer built in Go. Keyboard-driven navigation with real-time search/filter and cross-platform terminal support.

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
cd projects/015-twin-commander/3-development
go build -o twin-commander .
```

### Verify

```bash
./twin-commander
# Two panels should appear side by side showing the current directory
# Press q to exit
```

## Usage

Twin Commander is a full-screen TUI application with no command-line flags or arguments. Launch it from any directory:

```bash
./twin-commander
```

Both panels start at your current working directory. The left panel is active by default (cyan border).

### Keyboard Reference

| Key | Normal Mode | Filter Mode |
|-----|-------------|-------------|
| Up/Down | Move cursor | No effect |
| Enter | Open directory / no-op on file | Keep filter, exit filter mode |
| Backspace | Go to parent directory | Delete filter character |
| Tab | Switch active panel | No effect |
| q | Quit | Types into filter |
| Ctrl+C | Quit | Quit |
| r | Refresh directory listing | Types into filter |
| . | Toggle hidden files | Types into filter |
| / | Enter filter mode | Types into filter |
| Escape | No effect | Clear filter, exit filter mode |

### Navigation

Navigate into a directory by selecting it and pressing Enter. Go back to the parent directory with Backspace. After going up, the cursor lands on the directory you came from.

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

### Example 2: Finding Files with Filter

Press `/` to enter filter mode. Type `test` to show only files containing "test" in their name. Press Enter to keep the filter active while you browse. The filter clears automatically when you navigate into a subdirectory.

### Example 3: Viewing Hidden Files

Press `.` to reveal dotfiles like `.git/`, `.gitignore`, `.env`. The status bar changes from `4 items, 12.3K` to `[H] 7 items, 16.5K`. Press `.` again to hide them.

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

## Running Tests

```bash
go test -v ./...
```

## Building

```bash
go build -o twin-commander .
```

Produces a single static binary with no runtime dependencies.
