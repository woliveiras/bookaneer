package database

import "strings"

// BuildWhereClause joins conditions into a WHERE clause, or returns an empty
// string when conditions is empty. This avoids repeating the same three-line
// pattern across every List() method.
//
//	where := database.BuildWhereClause(conditions)
//	query := "SELECT ... FROM books " + where
func BuildWhereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

// NormaliseSortDir converts a caller-supplied sort direction string into the
// SQL keyword "ASC" or "DESC". Any value other than "desc" (case-sensitive)
// returns "ASC".
func NormaliseSortDir(dir string) string {
	if dir == "desc" {
		return "DESC"
	}
	return "ASC"
}

// ClampLimit returns limit when it falls in (0, maxVal], otherwise defaultVal.
// Use this to enforce a safe page size with a sensible default.
func ClampLimit(limit, defaultVal, maxVal int) int {
	if limit > 0 && limit <= maxVal {
		return limit
	}
	return defaultVal
}
