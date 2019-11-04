package utils

import (
	"testing"

	"github.com/ccding/go-logging/logging"
)

const (
	openTag  = "<b>"
	closeTag = "</b>"
)

type processorTestCase struct {
	positive        []string
	negative        []string
	text            string
	mustFail        bool
	positiveMatches []string
	negativeMatches []string
	output          string
	weight int
}

type highlightTestCase struct {
	text      string
	matches   []TextMatch
	processed string
}

var (
	processorTestCases = []processorTestCase{
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
			output: "<b>Смутное время</b>! Призрак свободы на коне, кровь по <b>колено</b>, словно в каком-то диком сне. - <b>Смутные времена</b>",
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
			output: "<b>Енот</b> <b>выпил</b> <b>весь</b> <b>компот</b>",
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
		{
			positive: []string{
				"ум",
			},
			text:   "Системы умного дома",
			output: "Системы умного дома",
		},
		{
			positive: []string{
				"*ум",
			},
			text:   "Системы умного безумного дома",
			output: "Системы умного безумного дома",
		},
		{
			positive: []string{
				"*ум*",
			},
			positiveMatches:[]string{
				"умного",
				"безумного",
			},
			text:   "Системы умного безумного дома",
			output: "Системы <b>умного</b> <b>безумного</b> дома",
		},
		{
			positive: []string{
				"ум*",
			},
			positiveMatches: []string {
				"умного",
			},
			text:   "Системы умного безумного дома",
			output: "Системы <b>умного</b> безумного дома",
		},
		{
			positive: []string{
				"*ум",
			},
			positiveMatches:[]string{
				"неум",
			},
			text:   "Системы умного безумного неум дома",
			output: "Системы умного безумного <b>неум</b> дома",
		},
		{
			positive: []string{
				"*умн* дом*",
			},
			positiveMatches:[]string{
				"безумный дом",
			},
			text:   "Это безумный дом",
			output: "Это <b>безумный дом</b>",
		},
		{
			positive: []string{
				"*умн* дом*",
			},
			text:   "Это безумный недом",
			output: "Это безумный недом",
		},
		{
			positive: []string{
				"*умн* дом*@5",
			},
			positiveMatches:[]string{
				"безумный дом",
			},
			text:   "Это безумный дом",
			output: "Это <b>безумный дом</b>",
			weight: 5,
		},
		{
			positive: []string{
				"енот*@5",
				"ночь@2",
			},
			positiveMatches:[]string{
				"енота",
				"енотов",
				"ночь",
			},
			text:   "Два енота у пруда водку пили до утра. Всю ночь они орали, а потом менты забрали. 'Про енотов'",
			output: "Два <b>енота</b> у пруда водку пили до утра. Всю <b>ночь</b> они орали, а потом менты забрали. 'Про <b>енотов</b>",
			weight: 12,
		},
	}

	highlightTestCases = []highlightTestCase{
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 7, End: 9},
				{Begin: 21, End: 25},
			},
			processed: "Racoon <b>is</b> the mammal-<b>size</b> animal",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 21, End: 25},
				{Begin: 7, End: 9},
			},
			processed: "Racoon <b>is</b> the mammal-<b>size</b> animal",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 21, End: 25},
				{Begin: 26, End: 32},
				{Begin: 7, End: 9},
			},
			processed: "Racoon <b>is</b> the mammal-<b>size</b> <b>animal</b>",
		},
		{
			text:      "Racoon is the mammal-size animal",
			processed: "Racoon is the mammal-size animal",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 7, End: 25},
				{Begin: 10, End: 13},
				{Begin: 14, End: 20},
			},
			processed: "Racoon <b>is the mammal-size</b> animal",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 7, End: 25},
				{Begin: 10, End: 13},
				{Begin: 14, End: 25},
			},
			processed: "Racoon <b>is the mammal-size</b> animal",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 2, End: 9},
				{Begin: 21, End: 32},
				{Begin: 0, End: 6},
				{Begin: 14, End: 25},
			},
			processed: "<b>Racoon is</b> the <b>mammal-size animal</b>",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 0, End: 32},
			},
			processed: "<b>Racoon is the mammal-size animal</b>",
		},
		{
			text: "Racoon is the mammal-size animal",
			matches: []TextMatch{
				{Begin: 4, End: 13},
				{Begin: 0, End: 32},
				{Begin: 10, End: 29},
				{Begin: 2, End: 18},
				{Begin: 30, End: 32},
				{Begin: 15, End: 17},
				{Begin: 28, End: 32},
			},
			processed: "<b>Racoon is the mammal-size animal</b>",
		},
	}
)

func TestTextProcessor(t *testing.T) {
	log, _ := logging.SimpleLogger("test")
	for i, test := range processorTestCases {
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
		if !isResultEqual(test.text, pos.Matches, test.positiveMatches) {
			t.Errorf("Test %d: '%+v' != '%+v'", i+1, getStringSlice(test.text, pos.Matches), test.positiveMatches)
		}

		if !isResultEqual(test.text, neg.Matches, test.negativeMatches) {
			t.Errorf("Test %d: '%+v' != '%+v'", i+1, getStringSlice(test.text, neg.Matches), test.negativeMatches)
		}

		if test.weight == 0 {
			test.weight = len(pos.Matches)
		}

		if test.weight != pos.Weight {
			t.Errorf("Test %d: Weight not equal: %d != %d", i+1, test.weight, pos.Weight)
		}

		highlighted := Highlight(test.text, pos.Matches, openTag, closeTag)
		if highlighted != test.output {
			t.Errorf("Test %d: '%s' != '%s'", i+1, highlighted, test.output)
		}

	}
}

func TestHighlight(t *testing.T) {
	for i, test := range highlightTestCases {
		result := Highlight(test.text, test.matches, openTag, closeTag)
		if result != test.processed {
			t.Errorf("TestCase #%d: Results are not equal: '%s' != '%s'", i+1, result, test.processed)
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
