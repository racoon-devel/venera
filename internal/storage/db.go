package storage

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

var db *gorm.DB

func Connect(connectionString string) error {
	var err error
	db, err = gorm.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	return db.AutoMigrate(&types.TaskRecord{}, &types.PersonRecord{}).Error
}

func LoadTasks() []types.TaskRecord {
	result := make([]types.TaskRecord, 0)
	db.Find(&result)
	return result
}

func AppendTask(task *types.TaskRecord) error {
	return db.Create(task).Error
}

func UpdateTask(task *types.TaskRecord) {
	db.Save(task)
}

func DeleteTask(task *types.TaskRecord) {
	db.Unscoped().Delete(&types.PersonRecord{}, "task_id = ?", task.ID)
	db.Unscoped().Delete(task)
}

func AppendPerson(person *types.Person, taskID uint, provider string) (uint, error) {
	data, err := json.Marshal(person)
	if err != nil {
		return 0, err
	}

	record := types.PersonRecord{TaskID: taskID, Description: string(data), Rating: person.Rating, PersonID: provider + "." + person.UserID}
	err = db.Create(&record).Error
	return record.ID, err
}

func LoadPersons(taskID uint, ascending bool, limit uint, offset uint, favourite bool, rating uint) ([]types.PersonRecord, uint, error) {
	persons := make([]types.PersonRecord, 0)
	ctx := db

	if ascending {
		ctx = ctx.Order("rating asc, created_at")
	} else {
		ctx = ctx.Order("rating desc, created_at")
	}

	if favourite {
		ctx = ctx.Where("favourite = ?", favourite)
	}

	if rating > 0 {
		ctx = ctx.Where("rating = ?", rating)
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

func LoadPerson(personID uint) (*types.PersonRecord, error) {
	record := types.PersonRecord{}
	db.First(&record, personID)

	if record.Description == "" {
		return nil, fmt.Errorf("Result has been deleted")
	}

	err := json.Unmarshal([]byte(record.Description), &record.Person)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func DeletePerson(personID uint) {
	record := types.PersonRecord{}
	record.ID = personID
	db.Delete(&record)
}

func Favourite(personID uint) {
	record := types.PersonRecord{}
	record.ID = personID
	db.Model(&record).Update("favourite", true)
}

func SearchPerson(providerID, userID string) *types.PersonRecord {
	record := types.PersonRecord{}
	ID := providerID+"."+userID
	db.Model(&record).Where("person_id = ?", ID).First(&record)
	if record.PersonID != ID {
		return nil
	}

	err := json.Unmarshal([]byte(record.Description), &record.Person)
	if err != nil {
		return nil
	}

	return &record
}
