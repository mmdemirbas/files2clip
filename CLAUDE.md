# files2clip

A CLI tool that reads file paths from a text file, collects their contents, and copies the formatted result to the system clipboard.

## Project structure

```
cmd/files2clip/main.go            — CLI entry point (os.Exit(run()) pattern), flag parsing, orchestration
cmd/files2clip/beep_*.go          — platform-specific beep (darwin/windows/other build files)
internal/clipboard/clipboard.go   — system clipboard read/write (pbcopy/pbpaste/xclip/powershell)
internal/completion/completion.go — shell completion scripts (bash/zsh/fish)
internal/config/config.go         — config loading, size parsing, defaults
internal/fileutil/binary.go       — binary file detection (null-byte heuristic)
internal/ignore/ignore.go         — gitignore-style pattern matching
internal/pathutil/commondir.go    — longest common directory prefix
internal/pathutil/pathutil.go     — path file reading, file exclusion
internal/style/style.go           — terminal colors, icons, formatting
```

## Build & test

Uses [Task](https://taskfile.dev) (Taskfile.yml):

```sh
task build     # build for current platform
task test      # run all tests
task lint      # go vet
task bench     # run benchmarks
task dist      # cross-compile all platforms
task all       # test + lint + dist
task clean     # remove build artifacts
```

Set version at build time: `task build VERSION=v1.0.0`

Release: tag with `vX.Y.Z`, push tag — CI builds binaries via `.goreleaser.yml` and creates GitHub Release.

## CLI interface

```
files2clip [flags] <path>...            # direct paths as arguments
files2clip -f <paths-file> [flags]      # read paths from a file
files2clip --from-clipboard [flags]     # read paths from clipboard

  --version                  print version
  --verbose                  detailed output (config, file sizes)
  -f, --file <path>          read paths from a file (one per line)
  --from-clipboard           read paths from clipboard
  --full-paths               use absolute paths in output
  --include-binary           include binary files (skipped by default)
  -e, --exclude <pattern>    exclude pattern (gitignore-style, repeatable)
  --ignore-file <path>       gitignore-style file for excluding paths
  --max-file-size <size>     override max file size (e.g., 10MB)
  --max-total-size <size>    override max total size (e.g., 50MB)
  --max-files <n>            override max file count
  --completion <shell>       generate shell completion (bash, zsh, fish)
```

## Config system

Priority: hardcoded defaults -> config file -> CLI flags.

Config file at `os.UserConfigDir()/files2clip/config` (simple key=value format).
Defaults: max_file_size=10MB, max_total_size=50MB, max_files=1000.
Boolean keys (default false): full_paths, include_binary.
Path keys: ignore_file (gitignore-style exclusion file).

## Conventions

### Language & dependencies
- Go 1.24+, module path: `github.com/mmdemirbas/files2clip`
- No external dependencies — stdlib only
- Static binaries: `CGO_ENABLED=0`, stripped with `-ldflags="-s -w"`
- Version injected at build time via `-X main.version`

### Code style
- `cmd/<name>/main.go` is thin — parses flags, calls `run()`, and exits. Only `main()` decides when to exit.
- `internal/` enforces encapsulation — not part of public API
- Small, focused functions — each does one thing, testable in isolation
- Closures for local helpers — avoid polluting package namespace
- No dead code — unused types, parameters, functions get deleted, not commented out
- Comments explain the **why**, not the **what**
- Platform-specific code uses build files (`_darwin.go`, `_windows.go`, `_other.go`), not runtime.GOOS switches

### Error handling
- Never swallow errors silently — every error must be handled or explicitly documented as intentional
- Wrap errors with context: `fmt.Errorf("doing X for %q: %w", name, err)`
- Return errors from libraries — never `log.Fatal` outside of `main()`
- Errors go to stderr, data output goes to stdout
- Exit codes: 0 = success, 1 = error
- Size limits reported explicitly, never silently truncated

### Cross-platform
- macOS (pbcopy/pbpaste), Linux (wl-copy/wl-paste or xclip), Windows (powershell Set-Clipboard/Get-Clipboard)
- Linux clipboard: prefers Wayland when WAYLAND_DISPLAY is set, falls back to X11
- Use `filepath.Separator`, `path.Match` for glob patterns
- Line ending awareness (`\r\n` vs `\n`)

### Testing
- Table-driven tests with `testing.T` and descriptive name field
- Benchmarks use `b.Loop()` (Go 1.24+), report with `-benchmem`
- Fuzz tests for parsers and user input processors (`go test -fuzz`)
- `t.Helper()` on all test helper functions
- `t.TempDir()` for throwaway files, `t.Setenv()` for env-dependent tests
- Edge cases: empty input, nil, zero values, boundary values, malformed input, broken symlinks, permission errors
- CI runs with `-race` flag

### Runtime behavior
- Symlink loops detected via resolved-path tracking
- Binary files skipped by default (null-byte detection in first 512 bytes, same as Git)
- `NO_COLOR` support — respect the [standard](https://no-color.org/)

### Performance
- Profile first, optimize second — intuition about bottlenecks is usually wrong
- Zero-allocation hot paths where possible
- `strings.Builder` / `bytes.Buffer` instead of string concatenation
- No `fmt.Sprintf` in hot loops
- Pre-compute at parse/init time, not at call time
