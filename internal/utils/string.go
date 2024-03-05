package utils

import "strings"

// Check string is empty
func IsEmpty(s string) bool {
	return strings.Trim(s, " ") == ""
}
