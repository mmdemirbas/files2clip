package style

import (
	"fmt"
	"os"
)

// Icons used in output messages.
const (
	IconOK    = "✓"
	IconSkip  = "⊘"
	IconFail  = "✗"
	IconLimit = "⚠"
	IconInfo  = "ℹ"
	IconDone  = "✔"
)

// ANSI color codes.
const (
	reset  = "\033[0m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	bold   = "\033[1m"
)

var colorEnabled = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// Check if stderr is a terminal
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// OK formats a success message.
func OK(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("  %s%s %s%s", green, IconOK, msg, reset)
	}
	return fmt.Sprintf("  %s %s", IconOK, msg)
}

// Skip formats a skip/warning message.
func Skip(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("  %s%s %s%s", yellow, IconSkip, msg, reset)
	}
	return fmt.Sprintf("  %s %s", IconSkip, msg)
}

// Fail formats an error message.
func Fail(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("  %s%s %s%s", red, IconFail, msg, reset)
	}
	return fmt.Sprintf("  %s %s", IconFail, msg)
}

// Limit formats a limit-reached message.
func Limit(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("  %s%s %s%s", yellow, IconLimit, msg, reset)
	}
	return fmt.Sprintf("  %s %s", IconLimit, msg)
}

// Info formats an informational message.
func Info(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("%s%s %s%s", cyan, IconInfo, msg, reset)
	}
	return fmt.Sprintf("%s %s", IconInfo, msg)
}

// Done formats a completion summary.
func Done(msg string) string {
	if colorEnabled {
		return fmt.Sprintf("%s%s%s %s%s", bold, green, IconDone, msg, reset)
	}
	return fmt.Sprintf("%s %s", IconDone, msg)
}
