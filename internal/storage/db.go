package storage

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type Storage struct {
	db *gorm.DB
}

func Connect(connectionString string) (*Storage, error) {
	storage := &Storage{}
	var err error
	storage.db, err = gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	storage.db.AutoMigrate(&types.TaskRecord{}, &types.PersonRecord{})
	return storage, err
}

func (self *Storage) LoadTasks() []types.TaskRecord {
	result := make([]types.TaskRecord, 0)
	self.db.Find(&result)
	return result
}

func (self *Storage) AppendTask(task *types.TaskRecord) {
	self.db.Create(task)
}

func (self *Storage) UpdateTask(task *types.TaskRecord) {
	self.db.Save(task)
}

func (self *Storage) DeleteTask(task *types.TaskRecord) {
	self.db.Delete(task)
}

func (self *Storage) AppendPerson(person *types.Person, taskID uint) {
	data, err := json.Marshal(person)
	if err != nil {
		panic(err)
	}

	record := types.PersonRecord{TaskID: taskID, Description: string(data), Rating: person.Rating}
	self.db.Create(&record)
}
