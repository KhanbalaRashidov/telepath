package transport

import "strings"

// truncate shortens s to n runes with an ellipsis.
func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func join(parts []string, sep string) string { return strings.Join(parts, sep) }

// BaseName returns the last path element, handling both separators.
func BaseName(p string) string {
	p = strings.TrimRight(p, "/\\")
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// SafeDiffName turns a file path into a safe "<name>.diff" attachment filename.
func SafeDiffName(path string) string {
	name := BaseName(path)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
	if name == "" {
		name = "changes"
	}
	return name + ".diff"
}
