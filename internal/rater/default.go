package rater

import (
	"time"

	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	minBioLength = 40
	relevantDays = 30
)

type defaultRater struct {
	settings  *types.SearchSettings
	processor *utils.TextProcessor
	detector  *utils.FaceDetector
}

func (r *defaultRater) Init(log *logging.Logger, settings *types.SearchSettings) {
	r.settings = settings
	var err error
	r.detector, err = utils.NewFaceDetector(utils.Configuration.Other.Content + "/cascade/facefinder")
	if err != nil {
		panic(err)
	}

	r.processor, err = utils.NewTextProcessor(log, settings.Likes, settings.Dislikes)
	if err != nil {
		panic(err)
	}
}

func (r *defaultRater) hasPhoto(photos []string) bool {
	for _, url := range photos {
		photo, err := utils.HttpRequest(url)
		if err == nil {
			result, _ := r.detector.IsFacePresent(photo)
			if result {
				return true
			}
		}
	}

	return false
}

// <0 - dislike, =0 random, >1 - like and save
func (r *defaultRater) Rate(person *types.Person) int {
	var rating int

	if person.Photo == nil || len(person.Photo) == 0 {
		return -1
	}

	if !r.hasPhoto(person.Photo) {
		return -1
	}

	if len(person.Bio) > minBioLength {
		rating++

		matches, dismatches := r.processor.Process(person.Bio)
		person.BioMatches = matches

		rating += len(person.BioMatches)
		rating -= len(dismatches)
	}

	person.Rating = rating
	return rating
}

func (r defaultRater) IsRelevant(visitDate time.Time) bool {
	distance := time.Now().Sub(visitDate)
	return distance.Hours()*24 <= relevantDays
}
