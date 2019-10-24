package tinder

import (
	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	minBioLength = 40
)

type tinderRater struct {
	settings  *types.SearchSettings
	processor *utils.TextProcessor
	detector  *utils.FaceDetector
}

func (self *tinderRater) Init(log *logging.Logger, settings *types.SearchSettings) {
	self.settings = settings
	var err error
	self.detector, err = utils.NewFaceDetector(utils.Configuration.Other.Content + "/cascade/facefinder")
	if err != nil {
		panic(err)
	}

	self.processor, err = utils.NewTextProcessor(log, settings.Likes, settings.Dislikes)
	if err != nil {
		panic(err)
	}
}

// <0 - dislike, =0 random, >1 - like and save
func (self *tinderRater) Rate(person *types.Person) int {
	var rating int

	if person.Photo == nil || len(person.Photo) == 0 {
		return -1
	}

	hasPhoto := false
	for _, url := range person.Photo {
		photo, err := utils.HttpRequest(url)
		if err == nil {
			result, _ := self.detector.IsFacePresent(photo)
			if result {
				hasPhoto = true
				break
			}
		}
	}

	if !hasPhoto {
		return -1
	}

	if len(person.Bio) > minBioLength {
		rating++

		matches, dismatches := self.processor.Process(person.Bio)
		person.BioMatches = matches

		rating += len(person.BioMatches)
		rating -= len(dismatches)
	}

	person.Rating = rating
	return rating
}
