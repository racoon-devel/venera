package storage

import "github.com/jinzhu/gorm"

func RawDB() *gorm.DB {
	return db
}
