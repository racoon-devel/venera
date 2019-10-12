package tinder

import (
	"regexp"
	"strings"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

type tinderRater struct {
	settings *types.SearchSettings
	likes    []*regexp.Regexp
	dislikes []*regexp.Regexp
	detector *utils.FaceDetector
}

func (self *tinderRater) Init(settings *types.SearchSettings) {
	self.settings = settings
	self.likes = make([]*regexp.Regexp, 0)
	for _, like := range settings.Likes {
		expr := regexp.MustCompile("[,.:;(*\\s](" + strings.ToLower(like) + ")")
		self.likes = append(self.likes, expr)
	}

	self.dislikes = make([]*regexp.Regexp, 0)
	for _, dislike := range settings.Dislikes {
		expr := regexp.MustCompile("[,.:;(*\\s](" + strings.ToLower(dislike) + ")")
		self.dislikes = append(self.dislikes, expr)
	}

	var err error
	self.detector, err = utils.NewFaceDetector(utils.Configuration.Other.CascadePath)
	if err != nil {
		panic(err)
	}
}

func getMatches(text string, exprs []*regexp.Regexp) []types.TextMatch {
	result := make([]types.TextMatch, 0)

	for _, expr := range exprs {
		matches := expr.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			result = append(result, types.TextMatch{Begin: match[2], End: match[3]})
		}
	}

	return result
}

// <0 - dislike, =0 random, >1 - like and save
func (self *tinderRater) Rate(person *types.Person) int {
	var rating int

	if person.Photo == nil || len(person.Photo) == 0 {
		return -1
	}

	photo, err := utils.HttpRequest(person.Photo[0])
	if err == nil {
		result, _ := self.detector.IsFacePresent(photo)
		if !result {
			return -1
		}
	}

	if person.Bio != "" {
		rating++

		text := strings.ToLower(person.Bio)
		person.BioMatches = getMatches(text, self.likes)
		dismatches := getMatches(text, self.dislikes)

		rating += len(person.BioMatches)
		rating -= len(dismatches)
	}

	person.Rating = rating
	return rating
}
