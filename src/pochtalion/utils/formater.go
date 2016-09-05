package utils

import "strings"

func Format(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
}
