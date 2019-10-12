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
	TaskID     uint
	UserID     string
	Rating     int `sql:"index"`
	Name       string
	Bio        string
	BioMatches []TextMatch
	Photo      []string `gorm:"-"`
}
