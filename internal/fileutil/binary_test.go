package fileutil

import (
	"bytes"
	"testing"
)

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", nil, false},
		{"plain text", []byte("hello world\n"), false},
		{"utf8 text", []byte("héllo wörld\n"), false},
		{"json", []byte(`{"key": "value"}`), false},
		{"go source", []byte("package main\n\nfunc main() {}\n"), false},

		{"null byte at start", []byte{0, 'h', 'e', 'l', 'l', 'o'}, true},
		{"null byte in middle", []byte("hel\x00lo"), true},
		{"null byte at end", []byte("hello\x00"), true},
		{"all nulls", []byte{0, 0, 0, 0}, true},
		{"png header", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}, true},
		{"elf header", []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00}, true},

		// Null byte beyond 512 bytes should not trigger detection
		{"null after 512", append(bytes.Repeat([]byte("a"), 512), 0), false},

		// Null byte exactly at position 511 (last checked byte)
		{"null at 511", append(bytes.Repeat([]byte("a"), 511), 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBinary(tt.data)
			if got != tt.want {
				t.Errorf("IsBinary(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func BenchmarkIsBinaryText(b *testing.B) {
	data := bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog.\n"), 100)
	for b.Loop() {
		IsBinary(data)
	}
}

func BenchmarkIsBinaryBinary(b *testing.B) {
	data := bytes.Repeat([]byte{0xFF, 0x00, 0xAB}, 200)
	for b.Loop() {
		IsBinary(data)
	}
}
