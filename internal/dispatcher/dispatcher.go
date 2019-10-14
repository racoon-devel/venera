package dispatcher

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ccding/go-logging/logging"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/storage"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

var dispatcher struct {
	log   *logging.Logger
	db    *storage.Storage
	tasks map[uint]*Task
	mutex sync.Mutex
}

func Init(log *logging.Logger) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.log = log

	var err error
	dispatcher.db, err = storage.Connect(utils.Configuration.GetConnectionString())
	if err != nil {
		return err
	}

	if err := webui.LoadTemplates(); err != nil {
		return err
	}

	dispatcher.tasks = make(map[uint]*Task, 0)
	taskRecords := dispatcher.db.LoadTasks()
	providers := provider.All()

	for _, record := range taskRecords {
		provider := providers[record.Provider]
		session := provider.RestoreSearchSession(dispatcher.log, record.CurrentState)

		task := newTask(session, &record)
		dispatcher.tasks[task.ID] = task
		log.Debugf("Task %s:#%d mode: %d", task.Provider, task.ID, task.Mode)

		task.Execute()
	}

	dispatcher.log.Debugf("Dispatcher inited")

	return nil
}

func AppendTask(session types.SearchSession, provider string) {
	record := types.TaskRecord{CurrentState: session.SaveState(), Provider: provider, Mode: ModeActive}
	dispatcher.db.AppendTask(&record)
	task := newTask(session, &record)

	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.tasks[task.ID] = task
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
		dispatcher.db.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d stopped", taskID)
	})
}

func DeleteTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Stop()
		delete(dispatcher.tasks, taskID)
		dispatcher.db.DeleteTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d deleted", taskID)
	})
}

func SuspendTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Suspend()
		dispatcher.db.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d suspended", taskID)
	})
}

func RunTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Run()
		dispatcher.db.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d started", taskID)
	})
}
