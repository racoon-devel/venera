package types

import (
	"github.com/jinzhu/gorm"
)

type TaskRecord struct {
	gorm.Model
	CurrentState string
	Provider     string
	Mode         int
}

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
	UserID     string
	Rating     int
	Name       string
	Bio        string
	BioMatches []TextMatch
	Photo      []string
}

type PersonRecord struct {
	gorm.Model
	TaskID      uint
	Rating      int `sql:"index"`
	Description string
}
