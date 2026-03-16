package pathutil

// CommonDir finds the longest common prefix directory of the given absolute paths.
func CommonDir(absolutePaths []string) string {
	if len(absolutePaths) == 0 {
		return ""
	}

	firstPath := absolutePaths[0]

	var separator byte
	if len(firstPath) >= 2 && firstPath[1] == ':' {
		separator = '\\' // Windows path separator
	} else {
		separator = '/' // Unix path separator
	}

	lastSeparatorIndex := -1
	for chIndex := 0; chIndex < len(firstPath); chIndex++ {
		ch := firstPath[chIndex]
		for _, path := range absolutePaths[1:] {
			if len(path) <= chIndex || path[chIndex] != ch {
				return firstPath[:lastSeparatorIndex+1]
			}
		}
		if ch == separator {
			lastSeparatorIndex = chIndex
		}
	}

	return firstPath[:lastSeparatorIndex+1]
}
