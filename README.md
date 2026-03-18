# files2clip

[![CI](https://github.com/mmdemirbas/files2clip/actions/workflows/ci.yml/badge.svg)](https://github.com/mmdemirbas/files2clip/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/mmdemirbas/files2clip)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Collect file contents and copy the formatted result to the system clipboard.

Pass file and directory paths as arguments, from a file, or from the clipboard. files2clip gathers their contents and copies everything to the clipboard in a markdown-friendly format — ready to paste into an LLM, a document, or an issue.

## Install

```sh
go install github.com/mmdemirbas/files2clip/cmd/files2clip@latest
```

Or download a pre-built binary from [Releases](https://github.com/mmdemirbas/files2clip/releases).

## Usage

```sh
# Pass paths directly as arguments
files2clip src/main.go src/util.go docs/

# Read paths from a file (one per line, # for comments)
files2clip -f paths.txt

# Read paths from clipboard
files2clip --from-clipboard
```

Output:

```
  ✓ main.go
  ✓ util.go
  ✓ docs/guide.md

✔ 3/3 files copied, relative to /home/user/project/
```

The formatted content is now on your clipboard.

## Flags

| Flag | Description |
|------|-------------|
| `--version` | Print version and exit |
| `--verbose` | Show config values, file sizes, and detailed info |
| `-f`, `--file <path>` | Read paths from a file (one per line) |
| `--from-clipboard` | Read paths from the system clipboard |
| `--full-paths` | Use absolute paths in output instead of relative |
| `--include-binary` | Include binary files (skipped by default) |
| `-e`, `--exclude <pattern>` | Exclude files matching a gitignore-style pattern (repeatable) |
| `--ignore-file <path>` | Path to a gitignore-style file for excluding paths |
| `--max-file-size <size>` | Max individual file size, e.g., `10MB` |
| `--max-total-size <size>` | Max total content size, e.g., `50MB` |
| `--max-files <n>` | Max number of files to process |
| `--completion <shell>` | Generate shell completion (bash, zsh, fish) |

Input modes (`<path>...`, `-f`, `--from-clipboard`) are mutually exclusive.

### Exclude patterns

Exclude patterns follow gitignore syntax:

```sh
# Skip test files and the vendor directory
files2clip -e "*.test.go" -e "vendor/" src/

# Use an ignore file for persistent exclusions
files2clip --ignore-file .clipignore src/
```

Supported pattern syntax: `*`, `?`, `**`, `[abc]`, `[a-z]`, `[!x]`, `!` negation, `#` comments, trailing `/` for directories, leading `/` for anchoring.

## Config

An optional config file sets default values. Location varies by OS:

| OS | Path |
|----|------|
| Linux | `~/.config/files2clip/config` |
| macOS | `~/Library/Application Support/files2clip/config` |
| Windows | `%AppData%\files2clip\config` |

Format — one key per line, `#` for comments:

```
# Max size per individual file
max_file_size = 10MB

# Max total clipboard content size
max_total_size = 50MB

# Max number of files to process
max_files = 1000

# Use absolute paths instead of relative
full_paths = false

# Path to a gitignore-style exclusion file
ignore_file = ~/.config/files2clip/ignore

# Include binary files (skipped by default)
include_binary = false
```

Size values accept `KB`, `MB`, `GB` suffixes (case-insensitive). Plain numbers are bytes.

CLI flags override config file values, which override built-in defaults.

When a limit is reached, files2clip reports it explicitly:

```
  ⚠ max_files=1000 reached, stopping
  ⊘ large-file.bin — exceeds max file size (15.0 MB > 10.0 MB)
  ⊘ image.png — binary file
```

## Shell Completions

```sh
# Bash
files2clip --completion bash > /etc/bash_completion.d/files2clip
# or for current user only:
mkdir -p ~/.local/share/bash-completion/completions
files2clip --completion bash > ~/.local/share/bash-completion/completions/files2clip

# Zsh
files2clip --completion zsh > /usr/local/share/zsh/site-functions/_files2clip

# Fish
files2clip --completion fish > ~/.config/fish/completions/files2clip.fish
```

Or use `task install-completions` to auto-detect your shell.

## Requirements

- **macOS**: `pbcopy` / `pbpaste` (built-in)
- **Linux (X11)**: `xclip` (`sudo apt install xclip`)
- **Linux (Wayland)**: `wl-clipboard` (`sudo apt install wl-clipboard`) — preferred when `WAYLAND_DISPLAY` is set, falls back to `xclip`
- **Windows**: PowerShell (built-in)

## Build from source

Requires [Go](https://go.dev/) 1.24+ and [Task](https://taskfile.dev/).

```sh
task build          # build for current platform
task test           # run tests
task lint           # static analysis
task bench          # run benchmarks
task dist           # cross-compile all platforms
task all            # test + lint + dist
task clean          # remove build artifacts
task setup:unix     # install clipboard tools on Linux
```

Set version at build time:

```sh
task build VERSION=v1.0.0
```

## License

[MIT](LICENSE)
