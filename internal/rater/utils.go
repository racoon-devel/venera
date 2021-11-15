package rater

import (
	"math"

	"github.com/racoon-devel/venera/internal/types"
)

func passNext(next types.Rater, person *types.Person, rating int) int {
	if next != nil && rating >= 0 {
		nextRating := next.Rate(person)
		person.Rating = int(math.Ceil((float64(rating) + float64(nextRating)) / 2))
		return person.Rating
	}

	person.Rating = rating
	return rating
}

func passThreshold(next types.Rater, thresholdType int, thresholdValue int) int {
	if next != nil {
		nextThreshold := next.Threshold(thresholdType)
		return int(math.Ceil((float64(thresholdValue) + float64(nextThreshold)) / 2))
	}

	return thresholdValue
}

func passNeedPhotos(next types.Rater, needPhotos bool) bool {
	if next != nil {
		return needPhotos || next.NeedPhotos()
	}

	return needPhotos
}

func propagateClose(next types.Rater) {
	if next != nil {
		next.Close()
	}
}
