package dispatcher

import (
	"sort"

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
	tasks map[uint]Task
}

func Init(log *logging.Logger) error {
	dispatcher.log = log

	var err error
	dispatcher.db, err = gorm.Open("postgres", utils.Configuration.GetConnectionString())
	if err != nil {
		return err
	}

	dispatcher.db.AutoMigrate(&Task{})

	dispatcher.tasks = make(map[uint]Task, 0)

	tasks := make([]Task, 0)
	dispatcher.db.Find(&tasks)
	providers := provider.All()

	for _, task := range tasks {
		provider := providers[task.Provider]
		task.session = provider.RestoreSearchSession(task.CurrentState)
		dispatcher.tasks[task.ID] = task

		task.Run()
	}

	return nil
}

func NewTask(session types.SearchSession, provider string) {
	task := Task{}
	task.CurrentState = session.SaveState()
	task.Provider = provider

	dispatcher.db.Create(&task)

	dispatcher.tasks[task.ID] = task
	task.Run()
}

func Describe() []TaskInfo {
	tasks := make([]TaskInfo, 0)
	for _, task := range dispatcher.tasks {
		tasks = append(tasks, task.GetInfo())
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Remaining < tasks[j].Remaining })

	return tasks
}
