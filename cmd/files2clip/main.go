package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmdemirbas/files2clip/internal/clipboard"
	"github.com/mmdemirbas/files2clip/internal/completion"
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
	completionShell := flag.String("completion", "", "generate shell completion (bash, zsh, fish)")
	var excludes stringSliceFlag
	flag.Var(&excludes, "exclude", "exclude pattern (gitignore-style, repeatable)")
	flag.Var(&excludes, "e", "exclude pattern (shorthand)")

	flag.Usage = printUsage

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0
	}
	if *completionShell != "" {
		return runCompletion(*completionShell)
	}

	n := inputModeCount(flag.NArg() > 0, *inputFile != "", *fromClipboard)
	if n == 0 {
		flag.Usage()
		return 1
	}
	if n > 1 {
		fmt.Fprintln(os.Stderr, style.Fail("specify only one input mode: paths as arguments, --file, or --from-clipboard"))
		return 1
	}

	cfg := loadEffectiveConfig(*maxFileSizeStr, *maxTotalSizeStr, *maxFilesFlag, *fullPaths, *ignoreFile, *includeBinary)
	logVerboseConfig(cfg, *verbose)

	paths, inputLabel, ignoreMatcher, err := setupExecution(*fromClipboard, *inputFile, flag.Args(), cfg.IgnoreFile, excludes)
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(err.Error()))
		return 1
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s doesn't contain any paths.", inputLabel)))
		return 1
	}

	if *verbose {
		fmt.Fprintln(os.Stderr, style.Info(fmt.Sprintf("input: %s", inputLabel)))
	}

	absolutePaths := collectFiles(paths, *verbose, ignoreMatcher)
	ancestor := computeAncestor(absolutePaths, cfg.FullPaths)
	res := buildOutput(absolutePaths, cfg, ancestor, *verbose)

	if code, done := checkResults(absolutePaths, res.successCount); done {
		return code
	}

	if err := clipboard.Set(res.buf.Bytes()); err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("clipboard: %v", err)))
		return 1
	}

	beep()
	fmt.Printf("\n%s\n", style.Done(buildSummary(res.successCount, len(absolutePaths), ancestor, res.totalSize, *verbose, res.filesLimitHit, res.sizeLimitHit)))
	return 0
}

func inputModeCount(hasArgs, hasFile, hasClipboard bool) int {
	n := 0
	if hasArgs {
		n++
	}
	if hasFile {
		n++
	}
	if hasClipboard {
		n++
	}
	return n
}

func logVerboseConfig(cfg config.Config, verbose bool) {
	if verbose {
		fmt.Fprintln(os.Stderr, style.Info(fmt.Sprintf(
			"config: max_file_size=%s, max_total_size=%s, max_files=%d, full_paths=%v",
			config.FormatSize(cfg.MaxFileSize), config.FormatSize(cfg.MaxTotalSize),
			cfg.MaxFiles, cfg.FullPaths)))
	}
}

func setupExecution(fromClipboard bool, inputFile string, args []string, ignoreFile string, excludes stringSliceFlag) ([]string, string, *ignore.Matcher, error) {
	m, err := loadIgnoreMatcher(ignoreFile, excludes)
	if err != nil {
		return nil, "", nil, fmt.Errorf("ignore file: %w", err)
	}
	paths, label, err := readInputPaths(fromClipboard, inputFile, args)
	if err != nil {
		return nil, "", nil, err
	}
	return paths, label, m, nil
}

func checkResults(absolutePaths []string, successCount int) (code int, done bool) {
	if len(absolutePaths) == 0 {
		fmt.Fprintln(os.Stderr, style.Skip("No files found."))
		return 0, true
	}
	if successCount == 0 {
		fmt.Fprintln(os.Stderr, style.Fail("No files copied."))
		return 1, true
	}
	return 0, false
}

func runCompletion(shell string) int {
	script, err := completion.Generate(shell)
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(err.Error()))
		return 1
	}
	fmt.Print(script)
	return 0
}

func computeAncestor(absolutePaths []string, fullPaths bool) string {
	if fullPaths {
		return ""
	}
	if a := pathutil.CommonDir(absolutePaths); a != "/" {
		return a
	}
	return ""
}

func buildSummary(successCount, total int, ancestor string, totalSize int64, verbose, filesLimitHit, sizeLimitHit bool) string {
	s := fmt.Sprintf("%d/%d files copied", successCount, total)
	if ancestor != "" {
		s += fmt.Sprintf(", relative to %s", ancestor)
	}
	if verbose {
		s += fmt.Sprintf(" (total %s)", config.FormatSize(totalSize))
	}
	if filesLimitHit || sizeLimitHit {
		s += " [limited]"
	}
	return s
}

func printUsage() {
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
	printFlag("--completion <shell>", "generate shell completion (bash, zsh, fish)")
	fmt.Fprintf(os.Stderr, "\nConfig file: %s\n", configFileHint())
	fmt.Fprintf(os.Stderr, "  Supported keys: max_file_size, max_total_size, max_files,\n")
	fmt.Fprintf(os.Stderr, "                  full_paths, ignore_file, include_binary\n")
}

func printFlag(name, desc string) {
	fmt.Fprintf(os.Stderr, "  %-26s %s\n", name, desc)
}

func loadEffectiveConfig(maxFileSizeStr, maxTotalSizeStr string, maxFilesFlag int, fullPaths bool, ignoreFile string, includeBinary bool) config.Config {
	cfg := config.DefaultConfig()
	if cfgPath, err := config.ConfigFilePath(); err == nil {
		if fileCfg, err := config.LoadFromFile(cfgPath); err == nil {
			cfg = fileCfg
		} else if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("config file error: %v", err)))
		}
	}
	applyCLIOverrides(&cfg, maxFileSizeStr, maxTotalSizeStr, maxFilesFlag, fullPaths, ignoreFile, includeBinary)
	return cfg
}

func loadIgnoreMatcher(ignoreFile string, excludes stringSliceFlag) (*ignore.Matcher, error) {
	var m *ignore.Matcher
	if ignoreFile != "" {
		var err error
		m, err = ignore.LoadFile(ignoreFile)
		if err != nil {
			return nil, err
		}
	}
	if len(excludes) > 0 {
		cli := ignore.Parse(strings.Join(excludes, "\n"))
		m = ignore.Merge(m, cli)
	}
	return m, nil
}

func readInputPaths(fromClipboard bool, inputFile string, args []string) ([]string, string, error) {
	if fromClipboard {
		data, err := clipboard.Get()
		if err != nil {
			return nil, "", fmt.Errorf("read clipboard: %w", err)
		}
		return pathutil.ParsePaths(string(data)), "(clipboard)", nil
	}
	if inputFile != "" {
		paths, err := pathutil.ReadPathsFromFile(inputFile)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", inputFile, err)
		}
		return paths, inputFile, nil
	}
	return args, "(args)", nil
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

// outputResult holds the accumulated result of processing a list of files.
type outputResult struct {
	buf           bytes.Buffer
	successCount  int
	totalSize     int64
	filesLimitHit bool
	sizeLimitHit  bool
}

func buildOutput(absolutePaths []string, cfg config.Config, ancestor string, verbose bool) outputResult {
	var res outputResult
	for _, ap := range absolutePaths {
		if cfg.MaxFiles > 0 && res.successCount >= cfg.MaxFiles {
			fmt.Fprintln(os.Stderr, style.Limit(fmt.Sprintf("max_files=%d reached, stopping", cfg.MaxFiles)))
			res.filesLimitHit = true
			break
		}
		if processOneFile(ap, ancestor, cfg, verbose, &res) {
			break
		}
	}
	return res
}

func processOneFile(ap, ancestor string, cfg config.Config, verbose bool, res *outputResult) (limitHit bool) {
	displayPath := displayName(ap, ancestor)

	fi, err := os.Stat(ap)
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", displayPath, err)))
		return false
	}
	if sizeExceedsLimit(cfg, fi, displayPath) {
		return false
	}

	data, err := os.ReadFile(ap) //nolint:gosec // G304: reading user-specified file paths is the tool's purpose
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — file not found", displayPath)))
		} else {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", displayPath, err)))
		}
		return false
	}

	if isBinaryFile(cfg, data, displayPath) {
		return false
	}

	// Calculate entry size: "path:\n```\n" + data + "\n```"
	entrySize := int64(len(displayPath)) + 6 + int64(len(data)) + 4
	if cfg.MaxTotalSize > 0 && res.totalSize+entrySize > cfg.MaxTotalSize {
		fmt.Fprintln(os.Stderr, style.Limit(fmt.Sprintf(
			"max_total_size=%s reached, stopping", config.FormatSize(cfg.MaxTotalSize))))
		res.sizeLimitHit = true
		return true
	}

	if res.successCount > 0 {
		res.buf.WriteString("\n\n")
	}
	res.buf.WriteString(displayPath)
	res.buf.WriteString(":\n```\n")
	res.buf.Write(data)
	res.buf.WriteString("\n```")
	res.totalSize += entrySize
	res.successCount++

	if verbose {
		fmt.Println(style.OK(fmt.Sprintf("%s (%s)", displayPath, config.FormatSize(fi.Size()))))
	} else {
		fmt.Println(style.OK(displayPath))
	}
	return false
}

func sizeExceedsLimit(cfg config.Config, fi os.FileInfo, displayPath string) bool {
	if cfg.MaxFileSize > 0 && fi.Size() > cfg.MaxFileSize {
		fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf(
			"%s — exceeds max file size (%s > %s)",
			displayPath, config.FormatSize(fi.Size()), config.FormatSize(cfg.MaxFileSize))))
		return true
	}
	return false
}

func isBinaryFile(cfg config.Config, data []byte, displayPath string) bool {
	if !cfg.IncludeBinary && fileutil.IsBinary(data) {
		fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — binary file", displayPath)))
		return true
	}
	return false
}

func displayName(ap, ancestor string) string {
	if ancestor != "" {
		if r, err := filepath.Rel(ancestor, ap); err == nil {
			return r
		}
	}
	return ap
}

func collectFiles(paths []string, verbose bool, ignoreMatcher *ignore.Matcher) []string {
	visited := make(map[string]bool)
	var absolutePaths []string
	for _, p := range paths {
		addPath(p, visited, &absolutePaths, ignoreMatcher, verbose)
	}
	return absolutePaths
}

func addPath(p string, visited map[string]bool, result *[]string, m *ignore.Matcher, verbose bool) {
	abs, err := filepath.Abs(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("resolve path %s: %v", p, err)))
		return
	}

	fi, err := os.Lstat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — not found", abs)))
		} else {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", abs, err)))
		}
		return
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — cannot resolve symlink: %v", abs, err)))
		return
	}

	if visited[resolved] {
		if verbose {
			fmt.Fprintln(os.Stderr, style.Skip(fmt.Sprintf("%s — already visited", abs)))
		}
		return
	}

	if !fi.IsDir() {
		if !pathutil.IsExcluded(abs) && !isIgnored(m, abs, false) {
			visited[resolved] = true
			*result = append(*result, abs)
		}
		return
	}

	_ = filepath.WalkDir(abs, walkFunc(visited, result, m, verbose))
}

func dirEntry(m *ignore.Matcher, path string) error {
	if isIgnored(m, path, true) {
		return filepath.SkipDir
	}
	return nil
}

func walkFunc(visited map[string]bool, result *[]string, m *ignore.Matcher, verbose bool) func(string, os.DirEntry, error) error {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Fail(fmt.Sprintf("%s — %v", path, err)))
			return nil
		}
		if d.IsDir() {
			return dirEntry(m, path)
		}
		if pathutil.IsExcluded(path) || isIgnored(m, path, false) {
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
		*result = append(*result, path)
		return nil
	}
}

func isIgnored(m *ignore.Matcher, path string, isDir bool) bool {
	if m == nil {
		return false
	}
	// Match already handles unanchored patterns by checking basename internally,
	// so a single call with the full path is sufficient.
	return m.Match(path, isDir)
}

func configFileHint() string {
	if p, err := config.ConfigFilePath(); err == nil {
		return p
	}
	return "(could not determine path)"
}
