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
	task.wg.Add(1)

	task.log.Infof("Starting task %s:#%d...", task.Provider, task.ID)

	go func() {
		task.log.Infof("Task %s:#%d started", task.Provider, task.ID)
		task.session.Process(&ctx)
		task.wg.Done()
	}()
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
