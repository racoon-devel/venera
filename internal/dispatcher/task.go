package dispatcher

import (
	"time"

	"github.com/jinzhu/gorm"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type TaskMode int

const (
	ModeActive = iota
	ModeIdle
)

type Task struct {
	gorm.Model
	CurrentState string
	Provider     string
	Mode         TaskMode
	session      types.SearchSession `gorm:"-"`
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

func (ctx *Task) Execute() {
	if ctx.Mode == ModeActive {
		ctx.Run()
	}
}

func (ctx *Task) Run() {
	ctx.Mode = ModeActive
	ctx.session.Start()
}

func (ctx *Task) Suspend() {
	ctx.Mode = ModeIdle
	ctx.session.Stop()
}

func (ctx *Task) Stop() {
	ctx.Mode = ModeIdle
	ctx.session.Stop()
	ctx.session.Reset()
}
