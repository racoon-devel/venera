package utils

import (
	"fmt"
	"regexp"
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
		matches := expr.FindAllStringIndex(text, -1)
		for _, match := range matches {
			result = append(result, TextMatch{Begin: match[0], End: match[1]})
			self.log.Debugf("mactched '%s' of '%s'", text[match[0]:match[1]], expr.String())
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
	expr := `\B` + keyword
	expr = strings.ReplaceAll(expr, "*", `[\p{L}]*`)
	expr = strings.ReplaceAll(expr, " ", `[\s]+`)
	return regexp.Compile(expr)
}
