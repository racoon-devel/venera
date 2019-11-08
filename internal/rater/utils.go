package rater

import "racoondev.tk/gitea/racoon/venera/internal/types"

func passNext(next types.Rater, person *types.Person, rating int) (int, int) {
	extra := 0
	if rating > maxPartialRating {
		extra = rating - maxPartialRating
		rating = maxPartialRating
	}

	if next != nil && rating >= 0 {
		rate, ex := next.Rate(person)
		return rate + rating, extra + ex
	}

	return rating, extra
}

func propagateClose(next types.Rater) {
	if next != nil {
		next.Close()
	}
}
