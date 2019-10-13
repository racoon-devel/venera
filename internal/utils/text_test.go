package utils

import (
	"testing"

	"github.com/ccding/go-logging/logging"
)

type testData struct {
	positive        []string
	negative        []string
	text            string
	mustFail        bool
	positiveMatches []string
	negativeMatches []string
}

var (
	tests = []testData{
		{
			positive: []string{
				"смутн* врем*",
				"го рулит",
				"*лено",
			},
			negative: []string{
				"кровь",
				"по колено",
			},
			text:     "Смутное время! Призрак свободы на коне, кровь по колено, словно в каком-то диком сне. - Смутные времена",
			mustFail: false,
			positiveMatches: []string{
				"Смутное время",
				"Смутные времена",
				"колено",
			},
			negativeMatches: []string{
				"кровь",
				"по колено",
			},
		},
		{
			positive: []string{
				"*",
			},
			text:     "Енот выпил весь компот",
			mustFail: false,
			positiveMatches: []string{
				"Енот",
				"выпил",
				"весь",
				"компот",
			},
		},
		{
			positive: []string{
				"смутн, врем*",
			},
			mustFail: true,
		},
		{
			positive: []string{
				"смутн7 врем*",
			},
			mustFail: true,
		},
		{
			negative: []string{
				"смутн, врем*",
			},
			mustFail: true,
		},
	}
)

func TestTextProcessor(t *testing.T) {
	log, _ := logging.SimpleLogger("test")
	for i, test := range tests {
		proc, err := NewTextProcessor(log, test.positive, test.negative)
		if test.mustFail && err == nil {
			t.Errorf("Test %d must fail", i+1)
		}

		if !test.mustFail && err != nil {
			t.Errorf("Test %d failed: %+v", i+1, err)
		}

		if test.mustFail {
			continue
		}

		pos, neg := proc.Process(test.text)
		if !isResultEqual(test.text, pos, test.positiveMatches) {
			t.Errorf("Test %d: '%+v' != '%+v'", i+1, getStringSlice(test.text, pos), test.positiveMatches)
		}

		if !isResultEqual(test.text, neg, test.negativeMatches) {
			t.Errorf("Test %d: '%+v' != '%+v'", i+1, getStringSlice(test.text, neg), test.negativeMatches)
		}

	}
}

func getStringSlice(text string, matches []TextMatch) []string {
	result := make([]string, 0)
	for _, match := range matches {
		str := text[match.Begin:match.End]
		result = append(result, str)
	}

	return result
}

func isResultEqual(text string, matches []TextMatch, expected []string) bool {
	for _, match := range matches {
		str := text[match.Begin:match.End]
		found := false
		for _, val := range expected {
			if val == str {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
