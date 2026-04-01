# Quality Findings (2026-04-01)

Run: `task quality`

## Status: ALL RESOLVED ✓

All 19 issues fixed across 9 commits. `task quality` passes with 0 issues.

---

## Priority 1 — Quick fixes (test file permissions) ✓

| # | File | Issue | Fix |
|---|------|-------|-----|
| ✓ | internal/config/config_test.go | G306: WriteFile perm 0644 | Changed to 0600 |
| ✓ | internal/ignore/ignore_test.go | G306: WriteFile perm 0644 | Changed to 0600 |
| ✓ | internal/pathutil/pathutil_test.go | G306: WriteFile perm 0644 | Changed to 0600 |

## Priority 2 — Intentional patterns (nosec annotations) ✓

| # | File | Issue | Fix |
|---|------|-------|-----|
| ✓ | internal/clipboard/clipboard.go:17 | G204: subprocess with variable | `#nosec` — clipboard cmd is OS-selected constant |
| ✓ | internal/clipboard/clipboard.go:28 | G204: subprocess with variable | `#nosec` — clipboard cmd is OS-selected constant |
| ✓ | cmd/files2clip/main.go | G304: file inclusion via variable | `#nosec` — reading user-specified paths is the tool's purpose |
| ✓ | internal/config/config.go | G304: file inclusion via variable | `#nosec` — user config file |
| ✓ | internal/ignore/ignore.go | G304: file inclusion via variable | `#nosec` — user ignore file |
| ✓ | internal/pathutil/pathutil.go | G304: file inclusion via variable | `#nosec` — user paths file |

## Priority 3 — Production complexity (refactoring) ✓

| # | File | Issue | Fix |
|---|------|-------|-----|
| ✓ | cmd/files2clip/main.go | gocognit `run` 69→≤15 | Extracted: runCompletion, computeAncestor, buildSummary, printUsage, inputModeCount, logVerboseConfig, setupExecution, checkResults |
| ✓ | cmd/files2clip/main.go | gocognit `collectFiles` 46→≤15 | Extracted: addPath, walkFunc, dirEntry |
| ✓ | cmd/files2clip/main.go | cyclop `run` 16→≤10 | Further extraction of printUsage |
| ✓ | cmd/files2clip/main.go | cyclop+funlen `processOneFile` | Extracted: sizeExceedsLimit, isBinaryFile, displayName, buildOutput |
| ✓ | internal/config/config.go | gocognit `LoadFromFile` 18→≤15 | Extracted: applyConfigLine, setSize |
| ✓ | internal/ignore/ignore.go | gocognit `Parse` 16→≤15 | Extracted: parsePattern |
| ✓ | internal/ignore/ignore.go | gocognit `doMatchParts` 18→≤15 | Extracted: matchDoublestar |

## Priority 4 — Test complexity ✓

| # | File | Issue | Fix |
|---|------|-------|-----|
| ✓ | internal/config/config_test.go | gocognit `TestLoadFromFile` 66→≤15 | Table-driven test with struct equality; flatten error check |
| ✓ | internal/clipboard/clipboard_test.go | gocognit `testClipboardCmd` 36→≤15 | Extracted: checkDarwinCmd, checkLinuxCmd, checkWindowsCmd, checkXclipArgs, checkWlPasteArgs |
| ✓ | internal/clipboard/clipboard_test.go | gocognit `TestLinuxCmd` 35→≤15 | Extracted: checkLinuxXclipFallback, checkLinuxWaylandPreferred, checkWaylandMode |
| ✓ | internal/style/style_test.go | gocognit `TestFormatFunctions` 22→≤15 | Extracted: checkFormatNoColor, checkFormatColor |
| ✓ | internal/ignore/ignore_test.go | gocognit `TestMerge` 16→≤15 | Converted nil cases to table-driven loop |
| ✓ | internal/ignore/ignore_test.go | funlen `TestMatch` 101→≤80 | Split into TestMatch + TestMatchDoublestar + TestMatchSpecialSyntax; extracted runMatchTests helper |

## Remaining coverage gaps

- cmd/files2clip: 0.0% — main package orchestration has no integration tests
- internal/clipboard: 13.8% — OS clipboard commands untested in CI (environment constraint)
- Total coverage: 41.2% (was 43.4% before refactoring — the main.go refactoring added more code paths)
