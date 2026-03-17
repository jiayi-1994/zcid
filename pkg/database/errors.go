package database

import "strings"

// IsUniqueConstraintError checks whether an error is caused by a unique constraint
// violation (PostgreSQL "duplicate key" or SQLite "UNIQUE constraint").
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
