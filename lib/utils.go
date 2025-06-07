package utils

import "strings"

func ContainsAny(content string, words []string) bool {
	for _, word := range words {
		if strings.Contains(content, word) {
			return true
		}
	}

	return false
}
