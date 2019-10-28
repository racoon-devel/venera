package dispatcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"sort"
	"sync"
	"time"

	"github.com/ccding/go-logging/logging"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"racoondev.tk/gitea/racoon/venera/internal/provider"
	"racoondev.tk/gitea/racoon/venera/internal/storage"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

var dispatcher struct {
	log        *logging.Logger
	tasks      map[uint]*Task
	mutex      sync.Mutex
	httpServer http.Server

	nightMode        bool
	cancelNightTimer context.CancelFunc
}

func Initialize(log *logging.Logger) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.log = log

	provider.SetLogger(log)

	if err := webui.LoadTemplates(); err != nil {
		return err
	}

	dispatcher.tasks = make(map[uint]*Task, 0)
	taskRecords := storage.LoadTasks()
	providers := provider.All()

	dispatcher.nightMode = utils.IsNightNow()

	for _, record := range taskRecords {
		provider := providers[record.Provider]
		session := provider.RestoreSearchSession(record.CurrentState)

		task := newTask(session, &record)
		dispatcher.tasks[task.ID] = task
		log.Debugf("Task %s:#%d mode: %d", task.Provider, task.ID, task.Mode)

		if !dispatcher.nightMode {
			task.Execute()
		}
	}

	registerBotCommands()

	dispatcher.log.Debugf("Dispatcher inited")

	return nil
}

func checkNightMode(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Hour):
			handleNightMode()
		}
	}
}

func handleNightMode() {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.log.Debug("Check night mode")

	isNight := utils.IsNightNow()
	if isNight && !dispatcher.nightMode {
		dispatcher.log.Info("Enter to night mode")
		for _, task := range dispatcher.tasks {
			task.Shutdown()
		}
		dispatcher.nightMode = true
	}

	if !isNight && dispatcher.nightMode {
		dispatcher.log.Info("Escape from night mode")
		for _, task := range dispatcher.tasks {
			task.Execute()
		}

		dispatcher.nightMode = false
	}
}

func getTaskProvider(taskID uint) types.Provider {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	task, ok := dispatcher.tasks[taskID]
	if !ok {
		return nil
	}

	prov, err := provider.Get(task.Provider)
	if err != nil {
		return nil
	}

	return prov
}

func AppendTask(session types.SearchSession, provider string) {
	record := types.TaskRecord{CurrentState: session.SaveState(), Provider: provider, Mode: ModeActive}
	storage.AppendTask(&record)
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

func GetTaskInfo(taskID uint) (TaskInfo, error) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	task, ok := dispatcher.tasks[taskID]
	if !ok {
		return TaskInfo{}, fmt.Errorf("Task not found: %d", taskID)
	}

	return task.GetInfo(), nil
}

func taskAction(taskID uint, handler func(task *Task)) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	task, ok := dispatcher.tasks[taskID]
	if !ok {
		return fmt.Errorf("Task not found: %d", taskID)
	}

	if dispatcher.nightMode {
		return fmt.Errorf("Night mode enabled. Ignore task control action")
	}

	handler(task)
	return nil
}

func StopTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Stop()
		storage.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d stopped", taskID)
	})
}

func DeleteTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Stop()
		delete(dispatcher.tasks, taskID)
		storage.DeleteTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d deleted", taskID)
	})
}

func SuspendTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Suspend()
		storage.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d suspended", taskID)
	})
}

func RunTask(taskID uint) error {
	return taskAction(taskID, func(task *Task) {
		task.Run()
		storage.UpdateTask(&task.TaskRecord)
		dispatcher.log.Infof("Task #%d started", taskID)
	})
}

func Action(taskID uint, action string, args url.Values) error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	task, ok := dispatcher.tasks[taskID]
	if !ok {
		return fmt.Errorf("Task not found: %d", taskID)
	}

	return task.Action(action, args)
}
