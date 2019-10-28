package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ccding/go-logging/logging"
)

type TextProcessor struct {
	log      *logging.Logger
	positive []*regexp.Regexp
	negative []*regexp.Regexp
}

type TextMatch struct {
	Begin int
	End   int
}

func NewTextProcessor(log *logging.Logger, positiveKeywords []string, negativeKeywords []string) (*TextProcessor, error) {
	self := &TextProcessor{log: log}
	var err error

	if err := Validate(positiveKeywords); err != nil {
		return nil, err
	}

	if err := Validate(negativeKeywords); err != nil {
		return nil, err
	}

	self.positive, err = self.compileExpressions(positiveKeywords)
	if err != nil {
		return nil, err
	}

	self.negative, err = self.compileExpressions(negativeKeywords)
	if err != nil {
		return nil, err
	}

	return self, nil
}

func Validate(patterns []string) error {
	for _, pattern := range patterns {
		if err := validate(pattern); err != nil {
			return err
		}
	}

	return nil
}

func (self *TextProcessor) Process(text string) (positiveMatches []TextMatch, negativeMatches []TextMatch) {
	text = strings.ToLower(text)
	positiveMatches = self.process(text, self.positive)
	negativeMatches = self.process(text, self.negative)
	return
}

func (self *TextProcessor) process(text string, exprs []*regexp.Regexp) []TextMatch {
	result := make([]TextMatch, 0)
	for _, expr := range exprs {
		matches := expr.FindAllStringSubmatchIndex(" " + text + " ", -1)
		for _, match := range matches {
			begin := match[2]-1
			end := match[3]-1
			if begin == end {
				continue
			}

			result = append(result, TextMatch{Begin: begin, End: end})
			self.log.Debugf("mactched '%s' of '%s'", text[begin:end], expr.String())
		}
	}

	return result
}

func (self *TextProcessor) compileExpressions(keywords []string) ([]*regexp.Regexp, error) {
	expressions := make([]*regexp.Regexp, 0)
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		expr, err := compile(keyword)
		if err != nil {
			return nil, err
		}

		self.log.Debugf("compiled '%s' -> '%s'", keyword, expr.String())

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func validate(pattern string) error {
	if strings.TrimSpace(pattern) != pattern {
		return fmt.Errorf("'%s' must have not leading and trailing spaces", pattern)
	}

	matched, _ := regexp.MatchString(`[[:punct:]]|\d|\n|\t`, strings.ReplaceAll(pattern, "*", ""))
	if matched {
		return fmt.Errorf("'%s' must have not digits, puncts or other symbols special symbols except '*'", pattern)
	}

	return nil
}

func compile(keyword string) (*regexp.Regexp, error) {
	var expr string

	expr += `\P{L}`


	expr += "(" + keyword
	expr = strings.ReplaceAll(expr, "*", `[\p{L}]*`)
	expr = strings.ReplaceAll(expr, " ", `[\s]+`)
	expr += `)`

	if keyword[len(keyword)-1] != '*' {
		expr += `\P{L}`
	}

	return regexp.Compile(expr)
}

type point struct {
	pos  int
	open bool
}

func getIntersection(matches []TextMatch) []TextMatch {
	points := make([]point, 0, len(matches)*2)
	for _, match := range matches {
		points = append(points, point{pos: match.Begin, open: true},
			point{pos: match.End, open: false})
	}

	sort.SliceStable(points, func(i, j int) bool {
		return points[i].pos < points[j].pos
	})

	result := make([]TextMatch, 0)
	counter := 0
	current := TextMatch{}

	for _, point := range points {
		if counter == 0 {
			current.Begin = point.pos
		}

		if !point.open {
			counter--
			if counter == 0 {
				current.End = point.pos - 1
				result = append(result, current)
			}
		} else {
			counter++
		}
	}
	return result
}

func Highlight(text string, matches []TextMatch, begin string, end string) string {
	if len(matches) == 0 || len(text) == 0 {
		return text
	}

	m := getIntersection(matches)
	output := ""
	var prev int

	for _, point := range m {
		output += text[prev:point.Begin] + begin + text[point.Begin:point.End+1] + end
		prev = point.End + 1
	}

	if prev < len(text)-1 {
		output += text[prev:]
	}

	return output
}
