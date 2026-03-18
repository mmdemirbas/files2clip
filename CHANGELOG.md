# Changelog

## Unreleased

### Added
- Direct path arguments: `files2clip src/ lib/main.go` (paths as CLI args)
- `-f` / `--file` flag to read paths from a file
- `--from-clipboard` flag to read paths from the system clipboard
- `-e` / `--exclude` repeatable flag for inline gitignore-style exclusion patterns
- `--ignore-file` flag and `ignore_file` config key for gitignore-style exclusion files
- `--full-paths` flag and `full_paths` config key to use absolute paths in output
- `--include-binary` flag and `include_binary` config key to include binary files
- Binary file detection: files with null bytes in the first 512 bytes are skipped by default
- Wayland clipboard support (`wl-copy` / `wl-paste`) with automatic fallback to `xclip`
- Colored output with icons (✓ success, ⊘ skip, ✗ error, ⚠ limit, ✔ done)
- Respects `NO_COLOR` and `TERM=dumb` for color-free output
- Gitignore pattern support: `*`, `?`, `**`, `[...]`, `[!...]`, `!` negation, `#` comments
- Shell completions for bash, zsh, and fish (`--completion <shell>`)
- goreleaser configuration for automated releases
- GitHub Actions CI with race detection and coverage reporting

### Changed
- Input file is now a named flag (`-f`) instead of a positional argument
- Input modes (args, file, clipboard) are validated as mutually exclusive
- Help output formatted with aligned columns and grouped sections
- Status messages use emoji icons instead of text prefixes

## v0.1.0

Initial release.

- Read file paths from a text file
- Collect file contents into markdown-formatted clipboard output
- Config file support (`max_file_size`, `max_total_size`, `max_files`)
- Cross-platform clipboard: macOS (pbcopy), Linux (xclip), Windows (PowerShell)
- Symlink loop detection
- OS metadata file exclusion (.DS_Store, Thumbs.db, desktop.ini)
