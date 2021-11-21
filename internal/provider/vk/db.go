package vk

import (
	"github.com/racoon-devel/venera/internal/storage"
)

type vkUserRecord struct {
	UserID int `gorm:"primary key;not null"`
}

func dbAdd(userID int) {
	storage.RawDB().Create(&vkUserRecord{UserID: userID})
}

func dbFetchFirst() (int, bool) {
	u := &vkUserRecord{}
	err := storage.RawDB().First(u).Error
	if err != nil {
		return -1, false
	}
	return u.UserID, true
}

func dbContains(usedID int) bool {
	var record vkUserRecord
	storage.RawDB().Model(&record).Where("user_id = ?", usedID).First(&record)
	return record.UserID == usedID
}

func dbRemove(userID int) {
	storage.RawDB().Where("user_id = ?", userID).Delete(&vkUserRecord{})
}
