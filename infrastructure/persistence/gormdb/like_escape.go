package gormdb

import "strings"

// likeContainsPattern wraps s as a SQL LIKE pattern with % wildcards and escapes \, %, _.
func likeContainsPattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return "%" + s + "%"
}
