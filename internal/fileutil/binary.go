package fileutil

// IsBinary reports whether data looks like binary content.
// It checks for null bytes in the first 512 bytes, the same heuristic Git uses.
func IsBinary(data []byte) bool {
	for i := range min(len(data), 512) {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
