package migrator

import "strings"

// splitSQLStatements splits a migration script into statements at top-level semicolons.
// It skips semicolons inside single-quoted strings (” = escaped quote) and
// treats -- line comments and /* */ block comments as non-code.
func splitSQLStatements(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	lineComment := false
	blockDepth := 0
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		if lineComment {
			if c == '\n' {
				lineComment = false
				_, _ = b.WriteRune(c)
			}
			continue
		}

		if blockDepth > 0 {
			if c == '/' && next == '*' {
				blockDepth++
				_, _ = b.WriteRune(c)
				_, _ = b.WriteRune(next)
				i++
				continue
			}
			if c == '*' && next == '/' {
				blockDepth--
				_, _ = b.WriteRune(c)
				_, _ = b.WriteRune(next)
				i++
				continue
			}
			_, _ = b.WriteRune(c)
			continue
		}

		if !inSingle && !inDouble {
			if c == '-' && next == '-' {
				lineComment = true
				_, _ = b.WriteRune(c)
				_, _ = b.WriteRune(next)
				i++
				continue
			}
			if c == '/' && next == '*' {
				blockDepth = 1
				_, _ = b.WriteRune(c)
				_, _ = b.WriteRune(next)
				i++
				continue
			}
		}

		if c == '\'' && !inDouble {
			if inSingle && next == '\'' {
				_, _ = b.WriteRune(c)
				_, _ = b.WriteRune(next)
				i++
				continue
			}
			inSingle = !inSingle
			_, _ = b.WriteRune(c)
			continue
		}
		if c == '"' && !inSingle {
			inDouble = !inDouble
			_, _ = b.WriteRune(c)
			continue
		}

		if c == ';' && !inSingle && !inDouble {
			stmt := strings.TrimSpace(b.String())
			b.Reset()
			if stmt != "" {
				out = append(out, stmt)
			}
			continue
		}

		_, _ = b.WriteRune(c)
	}

	rest := strings.TrimSpace(b.String())
	if rest != "" {
		out = append(out, rest)
	}

	// Drop pure-comment / whitespace-only fragments (defensive).
	var filtered []string
	for _, st := range out {
		if !isEffectivelyEmpty(st) {
			filtered = append(filtered, st)
		}
	}
	return filtered
}

func isEffectivelyEmpty(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	// Strip line comments only lines for check.
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "--") {
			continue
		}
		return false
	}
	return true
}
