package dispatcher

import (
	"time"

	"github.com/jinzhu/gorm"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type Task struct {
	gorm.Model
	CurrentState string
	Provider     string
	session      types.SearchSession `gorm:"-"`
}

type TaskInfo struct {
	ID        uint
	Remaining time.Duration
	Status    int
	Provider  string
}

func (ctx Task) GetInfo() TaskInfo {
	ti := TaskInfo{Provider: ctx.Provider, ID: ctx.ID}
	ti.Remaining = time.Now().Sub(ctx.CreatedAt)
	return ti
}

func (ctx Task) Run() {

}
