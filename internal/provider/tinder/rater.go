package tinder

import (
	"racoondev.tk/gitea/racoon/tindergo"
)

// <0 - dislike, =0 random, >1 - like and save
func rate(person *tindergo.RecsCoreUser, settings *searchSettings) int {
	var rating int

	if person.Bio != "" {
		rating++
	}

	return rating
}
