package types

import "github.com/jinzhu/gorm"

type SearchSettings struct {
	AgeFrom  uint
	AgeTo    uint
	Keywords []string
	Likes    []string
	Dislikes []string
}

type TextMatch struct {
	Begin int
	End   int
}

type Person struct {
	gorm.Model
	ID         string
	Rating     int
	Name       string
	Bio        string
	BioMatches []TextMatch
}
