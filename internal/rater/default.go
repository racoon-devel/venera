package rater

import (
	"fmt"
	"math"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"

	"github.com/BurntSushi/toml"
)

type defaultConfig struct {
	MinBioLength int
	RelevantDays int

	VIPAccountWeight int
	BioMatchWeight   int
	BioPresentWeight int

	AlcoFactor  int
	SmokeFactor int
	BodyFactor  int

	LikeThreshold      int
	SuperLikeThreshold int
}

type defaultRater struct {
	configName string
	scores     int
	settings   *types.SearchSettings
	processor  *utils.TextProcessor
	detector   *utils.FaceDetector
	config     defaultConfig
	log        *logging.Logger
	nextRater  types.Rater
}

func (r *defaultRater) Init(log *logging.Logger, settings *types.SearchSettings) {
	r.log = log
	r.settings = settings
	var err error
	r.detector, err = utils.NewFaceDetector(utils.Configuration.Directories.Content + "/cascade/facefinder")
	if err != nil {
		panic(err)
	}

	r.processor, err = utils.NewTextProcessor(log, settings.Likes, settings.Dislikes)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("%s/configurations/default.%s.conf", utils.Configuration.Directories.Content, r.configName)
	_, err = toml.DecodeFile(path, &r.config)
	if err != nil {
		panic(err)
	}

	r.scores += 1 + r.processor.GetMatchCount() + 2*types.Positive + types.Thin

	log.Debugf("Rater configuration [ default ] : %+v, max scores = %d", r.config, r.scores)
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

func (r *defaultRater) Rate(person *types.Person) int {
	var score int

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
		score += r.config.BioPresentWeight

		matches, dismatches := r.processor.Process(person.Bio)
		person.BioMatches = matches.Matches

		score += matches.Weight * r.config.BioMatchWeight
		score -= dismatches.Weight * r.config.BioMatchWeight
	}

	if person.VIP {
		score += r.config.VIPAccountWeight
	}

	score += person.Alco * r.config.AlcoFactor
	score += person.Smoke * r.config.SmokeFactor
	score += person.Body * r.config.BodyFactor

	x := (float64(score) / float64(r.scores)) * 100
	y := -(100 / (0.05 * (x + 16))) + 120

	if x < 0 || y < 0 {
		person.Rating = IgnorePerson
		return person.Rating
	}

	rating := int(math.Ceil(y))
	if rating > 100 {
		rating = 100
	}

	if rating < 0 {
		rating = 0
	}

	return passNext(r.nextRater, person, rating)
}

func (r *defaultRater) Next(nextRater types.Rater) types.Rater {
	r.nextRater = nextRater
	return nextRater
}

func (r *defaultRater) Threshold(thresholdType int) int {
	threshold := r.config.LikeThreshold
	if thresholdType == types.SuperLikeThreshold {
		threshold = r.config.SuperLikeThreshold
	}

	return passThreshold(r.nextRater, thresholdType, threshold)
}

func (r *defaultRater) NeedPhotos() bool {
	return passNeedPhotos(r.nextRater, false)
}

func (r *defaultRater) Close() {
	propagateClose(r.nextRater)
}
