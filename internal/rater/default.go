package rater

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"time"

	"github.com/BurntSushi/toml"
)

type defaultConfig struct {
	MinBioLength int
	RelevantDays int

	VIPAccountWeight int
	BioMatchWeight   int
	BioPresentWeight  int

	AlcoFactor  int
	SmokeFactor int
	BodyFactor  int
}

type defaultRater struct {
	configName string
	settings   *types.SearchSettings
	processor  *utils.TextProcessor
	detector   *utils.FaceDetector
	config     defaultConfig
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

	path := fmt.Sprintf("%s/configurations/default.%s.conf", utils.Configuration.Other.Content, r.configName)
	_, err = toml.DecodeFile(path, &r.config)
	if err != nil {
		panic(err)
	}

	log.Debugf("Rater configuration [ default ] : %+v", r.config)
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

	if person.Photo == nil || len(person.Photo) == 0 || person.Body == types.Fat {
		person.Rating = IgnorePerson
		return person.Rating
	}

	if !r.hasPhoto(person.Photo) {
		person.Rating = IgnorePerson
		return person.Rating
	}

	if !person.VisitTime.IsZero() {
		distance := time.Now().Sub(person.VisitTime)
		if distance.Hours()/24 > float64(r.config.RelevantDays) {
			person.Rating = IgnorePerson
			return person.Rating
		}
	}

	if len(person.Bio) > r.config.MinBioLength {
		rating += r.config.BioPresentWeight

		matches, dismatches := r.processor.Process(person.Bio)
		person.BioMatches = matches.Matches

		rating += matches.Weight * r.config.BioMatchWeight
		rating -= dismatches.Weight * r.config.BioMatchWeight
	}

	if person.VIP {
		rating += r.config.VIPAccountWeight
	}

	rating += person.Alco * r.config.AlcoFactor
	rating += person.Smoke * r.config.SmokeFactor
	rating += person.Body * r.config.BodyFactor

	person.Rating = rating
	return rating
}
