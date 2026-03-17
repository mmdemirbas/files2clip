package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mmdemirbas/files2clip/internal/clipboard"
	"github.com/mmdemirbas/files2clip/internal/config"
	"github.com/mmdemirbas/files2clip/internal/fileutil"
	"github.com/mmdemirbas/files2clip/internal/ignore"
	"github.com/mmdemirbas/files2clip/internal/pathutil"
	"github.com/mmdemirbas/files2clip/internal/style"
)

var version = "dev"

func main() {
	os.Exit(run())
}

// stringSliceFlag implements flag.Value for repeatable string flags.
type stringSliceFlag []string

func (f *stringSliceFlag) String() string { return strings.Join(*f, ", ") }
func (f *stringSliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func run() int {
	showVersion := flag.Bool("version", false, "print version and exit")
	verbose := flag.Bool("verbose", false, "show detailed processing info")
	maxFileSizeStr := flag.String("max-file-size", "", "max individual file size (e.g., 10MB)")
	maxTotalSizeStr := flag.String("max-total-size", "", "max total content size (e.g., 50MB)")
	maxFilesFlag := flag.Int("max-files", 0, "max number of files to process")
	fullPaths := flag.Bool("full-paths", false, "use absolute paths in output")
	ignoreFile := flag.String("ignore-file", "", "gitignore-style file for excluding paths")
	fromClipboard := flag.Bool("from-clipboard", false, "read paths from clipboard")
	inputFile := flag.String("file", "", "read paths from a file (one path per line)")
	flag.StringVar(inputFile, "f", "", "read paths from a file (shorthand)")
	includeBinary := flag.Bool("include-binary", false, "include binary files (skipped by default)")
	var excludes stringSliceFlag
	flag.Var(&excludes, "exclude", "exclude pattern (gitignore-style, repeatable)")
	flag.Var(&excludes, "e", "exclude pattern (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: files2clip [flags] <path>...\n")
		fmt.Fprintf(os.Stderr, "       files2clip -f <paths-file> [flags]\n")
		fmt.Fprintf(os.Stderr, "       files2clip --from-clipboard [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Collects file contents and copies the formatted result to the clipboard.\n\n")
		fmt.Fprintf(os.Stderr, "Input modes (mutually exclusive):\n")
		printFlag("<path>...", "file/directory paths as arguments (default)")
		printFlag("-f, --file <path>", "read paths from a file (one per line)")
		printFlag("--from-clipboard", "read paths from the system clipboard")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		printFlag("--version", "print version and exit")
		printFlag("--verbose", "show detailed processing info")
		printFlag("--full-paths", "use absolute paths in output")
		printFlag("--include-binary", "include binary files (skipped by default)")
		printFlag("-e, --exclude <pattern>", "exclude pattern (gitignore-style, repeatable)")
		printFlag("--ignore-file <path>", "gitignore-style file for excluding paths")
		printFlag("--max-file-size <size>", "max individual file size (e.g., 10MB)")
		printFlag("--max-total-size <size>", "max total content size (e.g., 50MB)")
		printFlag("--max-files <n>", "max number of files to process")
		fmt.Fprintf(os.Stderr, "\nConfig file: %s\n", configFileHint())
		fmt.Fprintf(os.Stderr, "  Supported keys: max_file_size, max_total_size, max_files,\n")
		fmt.Fprintf(os.Stderr, "                  full_paths, ignore_file, include_binary\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0
	}

	// Validate mutually exclusive input modes
	modeCount := 0
	if flag.NArg() > 0 {
		modeCount++
	}
	if *inputFile != "" {
		modeCount++
	}
	if *fromClipboard {
		modeCount++
	}
	if modeCount == 0 {
		flag.Usage()
		return 1
	}
	if modeCount > 1 {
		fmt.Fprintln(os.Stderr, style.Fail("specify only one input mode: paths as arguments, --file, or --from-clipboard"))
		return 1
	}

	// Load config: defaults → config file → CLI flags
	cfg := config.DefaultConfig()
	if cfgPath, err := config.ConfigFilePath(); err == nil {
		if fileCfg, err := config.LoadFromFile(cfgPath); err == nil {
			cfg = fileCfg
		} else if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("config file error: %v", err)))
		}
	}
	applyCLIOverrides(&cfg, *maxFileSizeStr, *maxTotalSizeStr, *maxFilesFlag, *fullPaths, *ignoreFile, *includeBinary)

	if *verbose {
		fmt.Fprintln(os.Stderr, style.Info(fmt.Sprintf(
			"config: max_file_size=%s, max_total_size=%s, max_files=%d, full_paths=%v",
			config.FormatSize(cfg.MaxFileSize), config.FormatSize(cfg.MaxTotalSize),
			cfg.MaxFiles, cfg.FullPaths)))
	}

	// Load ignore patterns (from file + inline --exclude flags)
	var ignoreMatcher *ignore.Matcher
	if cfg.IgnoreFile != "" {
		var err error
		ignoreMatcher, err = ignore.LoadFile(cfg.IgnoreFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("ignore file: %v", err)))
			return 1
		}
	}
	if len(excludes) > 0 {
		cliMatcher := ignore.Parse(strings.Join(excludes, "\n"))
		if ignoreMatcher != nil {
			ignoreMatcher = ignore.Merge(ignoreMatcher, cliMatcher)
		} else {
			ignoreMatcher = cliMatcher
		}
	}

	// Read paths from args, file, or clipboard
	var paths []string
	var inputLabel string
	switch {
	case *fromClipboard:
		data, err := clipboard.Get()
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("read clipboard: %v", err)))
			return 1
		}
		paths = pathutil.ParsePaths(string(data))
		inputLabel = "(clipboard)"
	case *inputFile != "":
		var err error
		paths, err = pathutil.ReadPathsFromFile(*inputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s: %v", *inputFile, err)))
			return 1
		}
		inputLabel = *inputFile
	default:
		paths = flag.Args()
		inputLabel = "(args)"
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s doesn't contain any paths.", inputLabel)))
		return 1
	}

	if *verbose {
		fmt.Fprintln(os.Stderr, style.Info(fmt.Sprintf("input: %s", inputLabel)))
	}

	absolutePaths := collectFiles(paths, *verbose, ignoreMatcher)

	ancestor := ""
	if !cfg.FullPaths {
		ancestor = pathutil.CommonDir(absolutePaths)
		if ancestor == "/" {
			ancestor = ""
		}
	}

	var contentList []string
	var totalSize int64
	successCount := 0
	filesLimitHit := false
	sizeLimitHit := false

	for _, ap := range absolutePaths {
		// Check max_files limit
		if cfg.MaxFiles > 0 && successCount >= cfg.MaxFiles {
			fmt.Fprintln(os.Stderr, style.Limit(fmt.Sprintf("max_files=%d reached, stopping", cfg.MaxFiles)))
			filesLimitHit = true
			break
		}

		displayPath := ap
		if ancestor != "" {
			if r, err := filepath.Rel(ancestor, ap); err == nil {
				displayPath = r
			}
		}

		// Check file size before reading
		fi, err := os.Stat(ap)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", displayPath, err)))
			continue
		}

		if cfg.MaxFileSize > 0 && fi.Size() > cfg.MaxFileSize {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf(
				"%s — exceeds max file size (%s > %s)",
				displayPath, config.FormatSize(fi.Size()), config.FormatSize(cfg.MaxFileSize))))
			continue
		}

		data, err := os.ReadFile(ap)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — file not found", displayPath)))
			} else {
				fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", displayPath, err)))
			}
			continue
		}

		if !cfg.IncludeBinary && fileutil.IsBinary(data) {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — binary file", displayPath)))
			continue
		}

		entry := fmt.Sprintf("%s:\n```\n%s\n```", displayPath, string(data))

		// Check max_total_size limit
		if cfg.MaxTotalSize > 0 && totalSize+int64(len(entry)) > cfg.MaxTotalSize {
			fmt.Fprintln(os.Stderr, style.Limit(fmt.Sprintf(
				"max_total_size=%s reached, stopping",
				config.FormatSize(cfg.MaxTotalSize))))
			sizeLimitHit = true
			break
		}

		contentList = append(contentList, entry)
		totalSize += int64(len(entry))
		successCount++

		if *verbose {
			fmt.Println(style.OK(fmt.Sprintf("%s (%s)", displayPath, config.FormatSize(fi.Size()))))
		} else {
			fmt.Println(style.OK(displayPath))
		}
	}

	total := len(absolutePaths)
	if total == 0 {
		fmt.Fprintln(os.Stderr, style.Skip("No files found."))
		return 0
	}
	if successCount == 0 {
		fmt.Fprintln(os.Stderr, style.Fail("No files copied."))
		return 1
	}

	formattedContent := strings.Join(contentList, "\n\n")
	if err := clipboard.Set([]byte(formattedContent)); err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("clipboard: %v", err)))
		return 1
	}

	beep()

	summary := fmt.Sprintf("%d/%d files copied", successCount, total)
	if ancestor != "" {
		summary += fmt.Sprintf(", relative to %s", ancestor)
	}
	if *verbose {
		summary += fmt.Sprintf(" (total %s)", config.FormatSize(totalSize))
	}
	if filesLimitHit || sizeLimitHit {
		summary += " [limited]"
	}
	fmt.Printf("\n%s\n", style.Done(summary))
	return 0
}

func printFlag(name, desc string) {
	fmt.Fprintf(os.Stderr, "  %-26s %s\n", name, desc)
}

func applyCLIOverrides(cfg *config.Config, maxFileSize, maxTotalSize string, maxFiles int, fullPaths bool, ignoreFile string, includeBinary bool) {
	if maxFileSize != "" {
		if n, err := config.ParseSize(maxFileSize); err == nil {
			cfg.MaxFileSize = n
		} else {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("invalid --max-file-size %q: %v", maxFileSize, err)))
		}
	}
	if maxTotalSize != "" {
		if n, err := config.ParseSize(maxTotalSize); err == nil {
			cfg.MaxTotalSize = n
		} else {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("invalid --max-total-size %q: %v", maxTotalSize, err)))
		}
	}
	if maxFiles > 0 {
		cfg.MaxFiles = maxFiles
	}
	if fullPaths {
		cfg.FullPaths = true
	}
	if ignoreFile != "" {
		cfg.IgnoreFile = ignoreFile
	}
	if includeBinary {
		cfg.IncludeBinary = true
	}
}

func collectFiles(paths []string, verbose bool, ignoreMatcher *ignore.Matcher) []string {
	visited := make(map[string]bool)
	var absolutePaths []string

	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("resolve path %s: %v", p, err)))
			continue
		}

		fi, err := os.Lstat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — not found", abs)))
			} else {
				fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", abs, err)))
			}
			continue
		}

		// Resolve symlinks for loop detection
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — cannot resolve symlink: %v", abs, err)))
			continue
		}

		if visited[resolved] {
			if verbose {
				fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — already visited", abs)))
			}
			continue
		}

		if !fi.IsDir() {
			if !pathutil.IsExcluded(abs) && !isIgnored(ignoreMatcher, abs, false) {
				visited[resolved] = true
				absolutePaths = append(absolutePaths, abs)
			}
			continue
		}

		filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", path, err)))
				return nil
			}

			if d.IsDir() {
				if isIgnored(ignoreMatcher, path, true) {
					return filepath.SkipDir
				}
				return nil
			}

			if pathutil.IsExcluded(path) || isIgnored(ignoreMatcher, path, false) {
				return nil
			}

			resolvedPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — cannot resolve symlink: %v", path, err)))
				return nil
			}

			if visited[resolvedPath] {
				if verbose {
					fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — already visited (symlink)", path)))
				}
				return nil
			}

			visited[resolvedPath] = true
			absolutePaths = append(absolutePaths, path)
			return nil
		})
	}
	return absolutePaths
}

func isIgnored(m *ignore.Matcher, path string, isDir bool) bool {
	if m == nil {
		return false
	}
	return m.Match(filepath.Base(path), isDir) || m.Match(path, isDir)
}

func configFileHint() string {
	if p, err := config.ConfigFilePath(); err == nil {
		return p
	}
	return "(could not determine path)"
}

func beep() {
	switch runtime.GOOS {
	case "windows":
		fmt.Print("\a")
	case "darwin":
		exec.Command("osascript", "-e", "beep").Run()
	}
}
