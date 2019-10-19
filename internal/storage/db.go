package storage

import (
	"encoding/json"
	"fmt"

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

func (self *Storage) AppendPerson(person *types.Person, taskID uint, provider string) {
	data, err := json.Marshal(person)
	if err != nil {
		panic(err)
	}

	record := types.PersonRecord{TaskID: taskID, Description: string(data), Rating: person.Rating, PersonID: provider + "." + person.UserID}
	self.db.Create(&record)
}

func (self *Storage) LoadPersons(taskID uint, ascending bool, limit uint, offset uint) ([]types.PersonRecord, uint, error) {
	persons := make([]types.PersonRecord, 0)
	ctx := self.db

	if ascending {
		ctx = ctx.Order("rating asc, created_at")
	} else {
		ctx = ctx.Order("rating desc, created_at")
	}

	if taskID != 0 {
		ctx = ctx.Where("task_id = ?", taskID)
	}

	var count uint
	ctx.Model(&types.PersonRecord{}).Count(&count)

	ctx = ctx.Limit(limit).Offset(offset)
	ctx.Find(&persons)

	for i, _ := range persons {
		err := json.Unmarshal([]byte(persons[i].Description), &persons[i].Person)
		if err != nil {
			return nil, 0, err
		}
	}

	return persons, count, nil
}

func (self *Storage) LoadPerson(personID uint) (*types.PersonRecord, error) {
	record := types.PersonRecord{}
	self.db.First(&record, personID)

	if record.Description == "" {
		return nil, fmt.Errorf("Result has been deleted")
	}

	err := json.Unmarshal([]byte(record.Description), &record.Person)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (self *Storage) DeletePerson(personID uint) {
	record := types.PersonRecord{}
	record.ID = personID
	self.db.Delete(&record)
}
