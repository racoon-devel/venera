package dispatcher

import (
	"context"
	"sync"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/jinzhu/gorm"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type TaskMode int

const (
	ModeIdle = iota
	ModeActive

	pollingInterval time.Duration = 2 * time.Minute
)

type Task struct {
	gorm.Model
	CurrentState string
	Provider     string
	Mode         TaskMode
	session      types.SearchSession `gorm:"-"`
	wg           sync.WaitGroup      `gorm:"-"`
	cancel       context.CancelFunc  `gorm:"-"`
	log          *logging.Logger     `gorm:"-"`
	timer        *time.Ticker        `gorm:"-"`
}

type TaskInfo struct {
	ID        uint
	Remaining time.Duration
	Status    types.SessionStatus
	Mode      TaskMode
	Provider  string
}

func (ctx Task) GetInfo() TaskInfo {
	ti := TaskInfo{Provider: ctx.Provider, ID: ctx.ID, Mode: ctx.Mode}
	ti.Remaining = time.Now().Sub(ctx.CreatedAt)
	ti.Status = ctx.session.Status()
	return ti
}

func (task *Task) Execute() {
	if task.Mode == ModeActive {
		task.start()
	}
}

func (task *Task) start() {
	task.Mode = ModeActive
	var ctx context.Context
	ctx, task.cancel = context.WithCancel(context.Background())
	task.wg.Add(2)

	task.log.Infof("Starting task %s:#%d...", task.Provider, task.ID)

	go func() {
		defer task.wg.Done()
		task.log.Infof("Task %s:#%d started", task.Provider, task.ID)
		task.session.Process(ctx)
	}()

	task.timer = time.NewTicker(pollingInterval)
	timerCtx, _ := context.WithCancel(ctx)

	go func(ctx context.Context) {
		defer task.wg.Done()
		for {
			select {
			case <-task.timer.C:
				task.poll()

			case <-ctx.Done():
				task.timer.Stop()
				return
			}
		}
	}(timerCtx)
}

func (task *Task) Run() {
	if task.Mode != ModeActive {
		task.start()
	}
}

func (task *Task) Suspend() {
	if task.Mode != ModeIdle {
		task.Mode = ModeIdle
		task.log.Infof("Stopping task %s:#%d...", task.Provider, task.ID)
		task.cancel()
		task.wg.Wait()
		task.log.Infof("Task %s:#%d stopped", task.Provider, task.ID)
	}
}

func (task *Task) Stop() {
	task.log.Debugf("task.Mode: %+v", task.Mode)

	if task.Mode != ModeIdle {
		task.Mode = ModeIdle
		task.log.Infof("Stopping task %s:#%d...", task.Provider, task.ID)
		task.cancel()
		task.wg.Wait()
		task.session.Reset()
		task.log.Infof("Task %s:#%d stopped", task.Provider, task.ID)
	}
}

func (task *Task) poll() {
	task.CurrentState = task.session.SaveState()
	dispatcher.db.Save(&task)
	matches := task.session.Results()
	for _, match := range matches {
		match.TaskID = task.ID
		dispatcher.db.Create(match)
	}

	dispatcher.log.Infof("polling: task %s:#%d got %d results", task.Provider, task.ID, len(matches))
}
