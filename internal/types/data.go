package types

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"racoondev.tk/gitea/racoon/venera/internal/utils"
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

type Person struct {
	UserID     string
	Rating     int
	Name       string
	Bio        string
	BioMatches []utils.TextMatch
	Photo      []string
}

type PersonRecord struct {
	gorm.Model
	TaskID      uint
	Rating      int `sql:"index"`
	Description string
}

func (self SearchSettings) Validate() error {
	if err := utils.Validate(self.Likes); err != nil {
		return err
	}

	if err := utils.Validate(self.Dislikes); err != nil {
		return err
	}

	if self.AgeTo < self.AgeFrom || self.AgeFrom < 18 || self.AgeTo < 18 {
		return fmt.Errorf("Invalid age: %d - %d", self.AgeFrom, self.AgeTo)
	}

	return nil
}
