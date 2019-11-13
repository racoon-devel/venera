package rater

import (
	"fmt"

	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const (
	IgnorePerson       = -100
	maxPartialRating   = 20
	LikeThreshold      = maxPartialRating / 2
	SuperLikeThreshold = maxPartialRating
)

type aggregator struct {
	raters    int
	nextRater types.Rater
}

type raterCreator func(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater

var (
	factoryMethods = map[string]raterCreator{
		"ml":         createMlRater,
		"default+ml": createCompositeRater,
		"default":    createDefaultRater,
	}
)

func NewRater(raterID string, configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	log.Debugf("Intancing rater '%s'", raterID)

	creator, ok := factoryMethods[raterID]
	if !ok {
		log.Warnf("Unknown rater: %s", raterID)
		return createDefaultRater(configuration, log, settings)
	}

	return creator(configuration, log, settings)
}

func GetRaters() []string {
	all := make([]string, 0)
	for k, _ := range factoryMethods {
		all = append(all, k)
	}

	return all
}

func createMlRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	rater := &mlRater{}
	rater.Init(log, settings)
	return rater
}

func createCompositeRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	def := &defaultRater{configName: configuration}
	def.Init(log, settings)

	ml := &mlRater{}
	ml.Init(log, settings)

	a := &aggregator{raters: 2}
	a.Next(def).Next(ml)

	return a
}

func createDefaultRater(configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	rater := &defaultRater{configName: configuration}
	rater.Init(log, settings)
	return rater
}

func (r *aggregator) Init(log *logging.Logger, settings *types.SearchSettings) {

}

func (r *aggregator) Rate(person *types.Person) (int, int) {
	rating, extra := r.nextRater.Rate(person)
	if rating >= maxPartialRating*r.raters {
		person.Rating = rating + extra
		return rating, extra
	}

	person.Rating = rating
	fmt.Printf("rating = %d\n", rating+extra)
	return rating, 0
}

func (r *aggregator) Next(nextRater types.Rater) types.Rater {
	r.nextRater = nextRater
	return nextRater
}

func (r *aggregator) Close() {
	propagateClose(r.nextRater)
}
