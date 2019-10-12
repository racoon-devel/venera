package dispatcher

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ccding/go-logging/logging"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var dispatcher struct {
	log   *logging.Logger
	db    *gorm.DB
	tasks map[uint]*Task
	mutex sync.Mutex
}

func Init(log *logging.Logger) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.log = log

	var err error
	dispatcher.db, err = gorm.Open("postgres", utils.Configuration.GetConnectionString())
	if err != nil {
		return err
	}

	dispatcher.db.AutoMigrate(&Task{})

	dispatcher.tasks = make(map[uint]*Task, 0)

	tasks := make([]Task, 0)
	dispatcher.db.Find(&tasks)
	providers := provider.All()

	for _, task := range tasks {
		provider := providers[task.Provider]
		task.session = provider.RestoreSearchSession(dispatcher.log, task.CurrentState)
		task.log = dispatcher.log
		dispatcher.tasks[task.ID] = &task
		log.Debugf("Task %s:#%d mode: %d", task.Provider, task.ID, task.Mode)

		task.Execute()
	}

	dispatcher.log.Debugf("Dispatcher inited")

	return nil
}

func NewTask(session types.SearchSession, provider string) {
	task := Task{log: dispatcher.log}
	task.CurrentState = session.SaveState()
	task.Provider = provider
	task.session = session
	task.Mode = ModeActive

	dispatcher.db.Create(&task)

	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.tasks[task.ID] = &task
	task.Execute()
}

func Describe() []TaskInfo {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	tasks := make([]TaskInfo, 0)
	for _, task := range dispatcher.tasks {
		tasks = append(tasks, task.GetInfo())
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Remaining < tasks[j].Remaining })

	return tasks
}

func taskAction(taskID uint, handler func(task *Task)) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	task, ok := dispatcher.tasks[taskID]
	if !ok {
		return fmt.Errorf("Task not found: %d", taskID)
	}

	handler(task)
	return nil
}

func StopTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Stop()
		dispatcher.db.Save(&task)
		dispatcher.log.Infof("Task #%d stopped", taskID)
	})
}

func DeleteTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Stop()
		delete(dispatcher.tasks, taskID)
		dispatcher.db.Delete(&task)
		dispatcher.log.Infof("Task #%d deleted", taskID)
	})
}

func SuspendTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Suspend()
		dispatcher.db.Save(&task)
		dispatcher.log.Infof("Task #%d suspended", taskID)
	})
}

func RunTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Run()
		dispatcher.db.Save(&task)
		dispatcher.log.Infof("Task #%d started", taskID)
	})
}
