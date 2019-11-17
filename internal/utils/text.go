package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ccding/go-logging/logging"
)

type matchPattern struct {
	expr *regexp.Regexp
	weight int
}

type TextProcessor struct {
	log      *logging.Logger
	positive []matchPattern
	negative []matchPattern
}

type TextMatch struct {
	Begin int
	End   int
}

type MatchResult struct {
	Matches []TextMatch
	Weight int
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

func (self *TextProcessor) Process(text string) (positiveMatches MatchResult, negativeMatches MatchResult) {
	text = strings.ToLower(text)
	positiveMatches = self.process(text, self.positive)
	negativeMatches = self.process(text, self.negative)
	return
}

func (self *TextProcessor) process(text string, patterns []matchPattern) MatchResult {
	result := MatchResult{Matches: []TextMatch{}}
	for _, pattern := range patterns {
		matches := pattern.expr.FindAllStringSubmatchIndex(" " + text + " ", -1)
		for _, match := range matches {
			begin := match[2]-1
			end := match[3]-1
			if begin == end {
				continue
			}

			result.Matches = append(result.Matches, TextMatch{Begin: begin, End: end})
			result.Weight += pattern.weight
			self.log.Debugf("matched '%s' of '%s'", text[begin:end], pattern.expr.String())
		}
	}

	return result
}

func (self *TextProcessor) compileExpressions(keywords []string) ([]matchPattern, error) {
	expressions := make([]matchPattern, 0)
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		expr, err := compile(keyword)
		if err != nil {
			return nil, err
		}

		self.log.Debugf("compiled '%s' -> '%s'", keyword, expr.expr.String())

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (self *TextProcessor) GetMatchCount() int {
	return len(self.positive)
}

func validate(pattern string) error {
	if strings.TrimSpace(pattern) != pattern {
		return fmt.Errorf("'%s' must have not leading and trailing spaces", pattern)
	}

	wexpr := regexp.MustCompile(`@[\d]+$`)
	strs := wexpr.FindAllString(pattern, 1)
	if len(strs) > 0 {
		pattern = strings.Replace(pattern, strs[0], "", 1)
	}

	matched, _ := regexp.MatchString(`[[:punct:]]|\d|\n|\t`, strings.ReplaceAll(pattern, "*", ""))
	if matched {
		return fmt.Errorf("'%s' must have not digits, puncts or other symbols special symbols except '*'", pattern)
	}

	return nil
}

func compile(keyword string) (matchPattern, error) {
	result := matchPattern{weight:1}

	wexpr := regexp.MustCompile(`@[\d]+$`)
	strs := wexpr.FindAllString(keyword, 1)
	if len(strs) > 0 {
		keyword = strings.Replace(keyword, strs[0], "", 1)
		weight, _ := strconv.ParseInt(strs[0][1:], 10, 16)
		result.weight = int(weight)
	}

	var expr string

	expr += `\P{L}`

	expr += "(" + keyword
	expr = strings.ReplaceAll(expr, "*", `[\p{L}]*`)
	expr = strings.ReplaceAll(expr, " ", `[\s]+`)
	expr += `)`

	if keyword[len(keyword)-1] != '*' {
		expr += `\P{L}`
	}

	var err error
	result.expr, err = regexp.Compile(expr)
	return result, err
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
