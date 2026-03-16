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
	"github.com/mmdemirbas/files2clip/internal/pathutil"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	showVersion := flag.Bool("version", false, "print version and exit")
	verbose := flag.Bool("verbose", false, "show detailed processing info")
	maxFileSizeStr := flag.String("max-file-size", "", "max individual file size (e.g., 10MB)")
	maxTotalSizeStr := flag.String("max-total-size", "", "max total content size (e.g., 50MB)")
	maxFilesFlag := flag.Int("max-files", 0, "max number of files to process")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: files2clip [flags] <paths-file>\n\n")
		fmt.Fprintf(os.Stderr, "Collects file contents and copies the formatted result to the clipboard.\n\n")
		fmt.Fprintf(os.Stderr, "The paths file contains one file or directory path per line.\n")
		fmt.Fprintf(os.Stderr, "Empty lines and lines starting with # are ignored.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfig file: %s\n", configFileHint())
		fmt.Fprintf(os.Stderr, "  Supported keys: max_file_size, max_total_size, max_files\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return 1
	}

	pathsFile := flag.Arg(0)

	// Load config: defaults → config file → CLI flags
	cfg := config.DefaultConfig()
	if cfgPath, err := config.ConfigFilePath(); err == nil {
		if fileCfg, err := config.LoadFromFile(cfgPath); err == nil {
			cfg = fileCfg
		} else if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: config file error: %v\n", err)
		}
	}
	applyCLIOverrides(&cfg, *maxFileSizeStr, *maxTotalSizeStr, *maxFilesFlag)

	if *verbose {
		fmt.Fprintf(os.Stderr, "config: max_file_size=%s, max_total_size=%s, max_files=%d\n",
			config.FormatSize(cfg.MaxFileSize), config.FormatSize(cfg.MaxTotalSize), cfg.MaxFiles)
		fmt.Fprintf(os.Stderr, "input:  %s\n\n", pathsFile)
	}

	paths, err := pathutil.ReadPathsFromFile(pathsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", pathsFile, err)
		return 1
	}
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "%s doesn't contain any paths.\n", pathsFile)
		return 1
	}

	absolutePaths := collectFiles(paths, *verbose)

	ancestor := pathutil.CommonDir(absolutePaths)
	if ancestor == "/" {
		ancestor = ""
	}

	var contentList []string
	var totalSize int64
	successCount := 0
	filesLimitHit := false
	sizeLimitHit := false

	for _, ap := range absolutePaths {
		// Check max_files limit
		if cfg.MaxFiles > 0 && successCount >= cfg.MaxFiles {
			fmt.Fprintf(os.Stderr, "LIMIT max_files=%d reached, stopping\n", cfg.MaxFiles)
			filesLimitHit = true
			break
		}

		rel := ap
		if ancestor != "" {
			if r, err := filepath.Rel(ancestor, ap); err == nil {
				rel = r
			}
		}

		// Check file size before reading
		fi, err := os.Stat(ap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s — %v\n", rel, err)
			continue
		}

		if cfg.MaxFileSize > 0 && fi.Size() > cfg.MaxFileSize {
			fmt.Fprintf(os.Stderr, "SKIP %s — exceeds max file size (%s > %s)\n",
				rel, config.FormatSize(fi.Size()), config.FormatSize(cfg.MaxFileSize))
			continue
		}

		data, err := os.ReadFile(ap)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "SKIP %s — file not found\n", rel)
			} else {
				fmt.Fprintf(os.Stderr, "FAIL %s — %v\n", rel, err)
			}
			continue
		}

		entry := fmt.Sprintf("%s:\n```\n%s\n```", rel, string(data))

		// Check max_total_size limit
		if cfg.MaxTotalSize > 0 && totalSize+int64(len(entry)) > cfg.MaxTotalSize {
			fmt.Fprintf(os.Stderr, "LIMIT max_total_size=%s reached, stopping\n",
				config.FormatSize(cfg.MaxTotalSize))
			sizeLimitHit = true
			break
		}

		contentList = append(contentList, entry)
		totalSize += int64(len(entry))
		successCount++

		if *verbose {
			fmt.Printf("  OK %s (%s)\n", rel, config.FormatSize(fi.Size()))
		} else {
			fmt.Printf("  OK %s\n", rel)
		}
	}

	total := len(absolutePaths)
	if total == 0 {
		fmt.Println("No files found.")
		return 0
	}
	if successCount == 0 {
		fmt.Fprintln(os.Stderr, "No files copied.")
		return 1
	}

	formattedContent := strings.Join(contentList, "\n\n")
	if err := clipboard.Set([]byte(formattedContent)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy to clipboard: %v\n", err)
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
	fmt.Printf("\n%s\n", summary)
	return 0
}

func applyCLIOverrides(cfg *config.Config, maxFileSize, maxTotalSize string, maxFiles int) {
	if maxFileSize != "" {
		if n, err := config.ParseSize(maxFileSize); err == nil {
			cfg.MaxFileSize = n
		} else {
			fmt.Fprintf(os.Stderr, "warning: invalid --max-file-size %q: %v\n", maxFileSize, err)
		}
	}
	if maxTotalSize != "" {
		if n, err := config.ParseSize(maxTotalSize); err == nil {
			cfg.MaxTotalSize = n
		} else {
			fmt.Fprintf(os.Stderr, "warning: invalid --max-total-size %q: %v\n", maxTotalSize, err)
		}
	}
	if maxFiles > 0 {
		cfg.MaxFiles = maxFiles
	}
}

func collectFiles(paths []string, verbose bool) []string {
	visited := make(map[string]bool)
	var absolutePaths []string

	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve path %s: %v\n", p, err)
			continue
		}

		fi, err := os.Lstat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "SKIP %s — not found\n", abs)
			} else {
				fmt.Fprintf(os.Stderr, "FAIL %s — %v\n", abs, err)
			}
			continue
		}

		// Resolve symlinks for loop detection
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s — cannot resolve symlink: %v\n", abs, err)
			continue
		}

		if visited[resolved] {
			if verbose {
				fmt.Fprintf(os.Stderr, "SKIP %s — already visited\n", abs)
			}
			continue
		}

		if !fi.IsDir() {
			if !pathutil.IsExcluded(abs) {
				visited[resolved] = true
				absolutePaths = append(absolutePaths, abs)
			}
			continue
		}

		filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "FAIL %s — %v\n", path, err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if pathutil.IsExcluded(path) {
				return nil
			}

			resolvedPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "FAIL %s — cannot resolve symlink: %v\n", path, err)
				return nil
			}

			if visited[resolvedPath] {
				if verbose {
					fmt.Fprintf(os.Stderr, "SKIP %s — already visited (symlink)\n", path)
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
