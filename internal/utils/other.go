package utils

import "strings"

// Bool2Int Type converter - C bool
var Bool2Int = map[bool]int{true: 1, false: 0}

func TrimArray(s []string) {
	for i := range s {
		s[i] = strings.Trim(s[i], " ")
	}
}
