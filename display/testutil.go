package display

import "strings"

// dedent removes any common leading whitespace from every line in text.
// It matches the Python textwrap.dedent algorithm: only a common prefix of identical whitespace is removed.
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	var margin string
	marginSet := false
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue // skip blank or whitespace-only lines
		}
		// Get leading whitespace
		prefix := line[:len(line)-len(trimmed)]
		if prefix == "" {
			margin = ""
			marginSet = true
			break // any non-blank line with no indent: margin is zero
		}
		if !marginSet {
			margin = prefix
			marginSet = true
		} else {
			// Find common prefix
			max := len(margin)
			if len(prefix) < max {
				max = len(prefix)
			}
			j := 0
			for ; j < max; j++ {
				if margin[j] != prefix[j] {
					break
				}
			}
			margin = margin[:j]
			if margin == "" {
				break
			}
		}
	}
	if margin == "" {
		return s // no margin, return original string
	}
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			lines[i] = "" // whitespace-only lines become empty
		} else if strings.HasPrefix(line, margin) {
			lines[i] = line[len(margin):]
		}
	}
	return strings.Join(lines, "\n")
}

// trimdedent trims only the leading and trailing newlines, leaving tabs and other whitespace alone, then dedents.
func trimdedent(s string) string {
	// Remove a single leading newline if present
	if len(s) > 0 && s[0] == '\n' {
		s = s[1:]
	}
	// Remove a single trailing newline if present
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return dedent(s)
}
