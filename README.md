# files2clip

Collect file contents and copy the formatted result to the system clipboard.

Reads file and directory paths from a text file, gathers their contents, and copies everything to the clipboard in a markdown-friendly format — ready to paste into an LLM, a document, or an issue.

## Install

```sh
go install github.com/mmdemirbas/files2clip/cmd/files2clip@latest
```

Or download a pre-built binary from [Releases](https://github.com/mmdemirbas/files2clip/releases).

## Usage

```sh
files2clip <paths-file>
```

The paths file contains one file or directory path per line. Empty lines and lines starting with `#` are ignored.

Example `paths.txt`:

```
# Project source files
src/main.go
src/util.go
docs/
```

```sh
files2clip paths.txt
```

Output:

```
  OK main.go
  OK util.go
  OK docs/guide.md

3/3 files copied, relative to /home/user/project/
```

The formatted content is now on your clipboard.

## Flags

| Flag | Description |
|------|-------------|
| `--help` | Show usage |
| `--version` | Print version and exit |
| `--verbose` | Show config values, file sizes, and detailed info |
| `--max-file-size` | Max individual file size, e.g., `10MB` (overrides config) |
| `--max-total-size` | Max total content size, e.g., `50MB` (overrides config) |
| `--max-files` | Max number of files to process (overrides config) |

## Config

An optional config file sets default limits. Location varies by OS:

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
```

Size values accept `KB`, `MB`, `GB` suffixes (case-insensitive). Plain numbers are bytes.

CLI flags override config file values, which override built-in defaults.

When a limit is reached, files2clip reports it explicitly:

```
LIMIT max_files=1000 reached, stopping
SKIP large-file.bin — exceeds max file size (15.0 MB > 10.0 MB)
```

## Requirements

- **macOS**: `pbcopy` (built-in)
- **Linux**: `xclip` (`sudo apt install xclip`)
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
task setup:unix     # install xclip on Linux
```

## License

[MIT](LICENSE)
