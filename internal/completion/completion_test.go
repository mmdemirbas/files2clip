package completion

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
		contain string
	}{
		{"bash", false, "complete -o default -F _files2clip files2clip"},
		{"zsh", false, "#compdef files2clip"},
		{"fish", false, "complete -c files2clip"},
		{"powershell", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got, err := Generate(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate(%q) error = %v, wantErr %v", tt.shell, err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(got, tt.contain) {
				t.Errorf("Generate(%q) missing %q", tt.shell, tt.contain)
			}
		})
	}
}

func TestCompletionScriptsContainAllFlags(t *testing.T) {
	flags := []string{
		"version", "verbose", "full-paths", "include-binary",
		"from-clipboard", "file", "exclude", "ignore-file",
		"max-file-size", "max-total-size", "max-files",
	}

	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			script, err := Generate(shell)
			if err != nil {
				t.Fatal(err)
			}
			for _, flag := range flags {
				if !strings.Contains(script, flag) {
					t.Errorf("%s completion missing flag %q", shell, flag)
				}
			}
		})
	}
}
