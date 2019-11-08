package rater

import (
	"fmt"
	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const (
	IgnorePerson = -100
	maxPartialRating = 20
)

type aggregator struct {
	raters int
	nextRater types.Rater
}

func NewRater(raterID string, configuration string, log *logging.Logger, settings *types.SearchSettings) types.Rater {
	switch raterID {
	case "ml":
		rater := &mlRater{}
		rater.Init(log, settings)
		return rater
	case "default+ml":
		def := &defaultRater{configName:configuration}
		def.Init(log, settings)

		ml := &mlRater{}
		ml.Init(log, settings)

		a := &aggregator{raters:2}
		a.Next(def).Next(ml)

		return a

	default:
		rater := &defaultRater{configName: configuration}
		rater.Init(log, settings)
		return rater
	}
}

func (r *aggregator) Init(log *logging.Logger, settings *types.SearchSettings) {

}

func (r *aggregator) Rate(person *types.Person) (int, int) {
	rating, extra := r.nextRater.Rate(person)
	if rating >= maxPartialRating * r.raters {
		person.Rating = rating + extra
		fmt.Printf("aggregated rating = %d\n", rating + extra)
		return rating, extra
	}

	person.Rating = rating
	fmt.Printf("rating = %d\n", rating + extra)
	return rating, 0
}

func (r *aggregator) Next(nextRater types.Rater) types.Rater {
	r.nextRater = nextRater
	return nextRater
}
