package fileutil

import "bytes"

// IsBinary reports whether data looks like binary content.
// It checks for null bytes in the first 512 bytes, the same heuristic Git uses.
func IsBinary(data []byte) bool {
	n := min(len(data), 512)
	return bytes.IndexByte(data[:n], 0) >= 0
}
