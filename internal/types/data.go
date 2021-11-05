package types

import (
	"fmt"
	"html/template"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/racoon-devel/venera/internal/utils"
)

const MyAge = 28

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
	Rater    string
}

const (
	Negative   = -1
	NotDefined = 0
	Neutral    = 1
	Positive   = 2
)

const (
	Fat   = -1
	Sport = 1
	Thin  = 2
)

type Person struct {
	UserID     string
	Rating     int
	Name       string
	Bio        string
	BioMatches []utils.TextMatch
	Photo      []string
	Job        string
	School     string
	Age        uint
	VIP        bool
	Alco       int
	Smoke      int
	Body       int
	VisitTime  time.Time
	Link       string
}

type Action struct {
	Link    template.URL
	Command string
	Title   string
}

type PersonRecord struct {
	gorm.Model
	TaskID      uint
	PersonID    string `gorm:"unique;not null"`
	Rating      int    `sql:"index"`
	Description string
	Favourite   bool
	Person      Person `gorm:"-"`
}

func (self SearchSettings) Validate() error {
	if err := utils.Validate(self.Likes); err != nil {
		return err
	}

	if err := utils.Validate(self.Dislikes); err != nil {
		return err
	}

	if self.AgeTo != 0 && self.AgeFrom != 0 {
		if self.AgeTo < self.AgeFrom || self.AgeFrom < 18 || self.AgeTo < 18 {
			return fmt.Errorf("Invalid age: %d - %d", self.AgeFrom, self.AgeTo)
		}
	}

	return nil
}
