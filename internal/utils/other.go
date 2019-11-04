package utils

import (
	"math/rand"
	"strings"
)

const (
	randomDistanceCo float32 = 0.005
)

// Bool2Int Type converter - C bool
var Bool2Int = map[bool]int{true: 1, false: 0}

func TrimArray(s []string) {
	for i := range s {
		s[i] = strings.Trim(s[i], " ")
	}
}

func RandomCoordinate(base float32) float32 {
	sign := rand.Int()%2 == 0
	if sign {
		return base + rand.Float32()*randomDistanceCo
	} else {
		return base - rand.Float32()*randomDistanceCo
	}
}

func ListToString(list []string) string {
	var result string
	for _, item := range list {
		result += item + ","
	}

	if len(result) != 0 {
		result = strings.TrimSuffix(result, result[len(result)-1:])
	}

	return result
}
