package dispatcher

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/racoon-devel/venera/internal/bot"

	"github.com/ccding/go-logging/logging"

	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/webui"
)

type TaskMode int

const (
	ModeIdle = iota
	ModeActive

	pollingInterval time.Duration = 1 * time.Minute
	statInterval    time.Duration = 1 * time.Hour
)

type Task struct {
	types.TaskRecord
	session      types.SearchSession
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	log          *logging.Logger
	timer        *time.Ticker
	lastStatTime time.Time
}

type TaskInfo struct {
	ID        uint
	Remaining time.Duration
	Status    types.SessionStatus
	Mode      TaskMode
	Provider  string
}

func newTask(session types.SearchSession, record *types.TaskRecord) *Task {
	task := &Task{session: session, log: dispatcher.log}
	task.TaskRecord = *record
	return task
}

func (ctx Task) GetInfo() TaskInfo {
	ti := TaskInfo{Provider: ctx.Provider, ID: ctx.ID, Mode: TaskMode(ctx.Mode)}
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
		task.session.Process(ctx, task.ID)
	}()

	task.timer = time.NewTicker(pollingInterval)
	task.lastStatTime = time.Now()
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
		task.poll()
		task.log.Infof("Task %s:#%d stopped", task.Provider, task.ID)
	}
}

func (task *Task) Shutdown() {
	if task.Mode != ModeIdle {
		task.log.Infof("Shutdowning task %s:#%d...", task.Provider, task.ID)
		task.cancel()
		task.wg.Wait()
		task.poll()
	}
}

func (task *Task) Stop() {
	if task.Mode != ModeIdle {
		task.Suspend()
		task.session.Reset()

		task.CurrentState = task.session.SaveState()
		storage.UpdateTask(&task.TaskRecord)
	}
}

func (task *Task) poll() {
	task.CurrentState = task.session.SaveState()
	storage.UpdateTask(&task.TaskRecord)
	task.session.Poll()

	if now := time.Now(); now.Sub(task.lastStatTime) >= statInterval {
		task.lastStatTime = now
		task.SendStat()
	}

	if err := task.session.GetLastError(); err != nil {
		msg := bot.Message{Content: fmt.Sprintf("Task #%d [ %s ] raised error: %+v",
			task.ID, task.Provider, err)}
		bot.Post(&msg)
	}
}

func (task *Task) WebUpdate(w http.ResponseWriter, r *http.Request) (bool, error) {
	log.Debugf("Editing task #%d", task.ID)
	updated, err := task.session.Update(w, r)
	if updated {
		task.CurrentState = task.session.SaveState()
		storage.UpdateTask(&task.TaskRecord)

		if task.Mode == ModeActive {
			task.Suspend()
			task.Run()
		}

		return true, nil
	}

	return false, err
}

func (task *Task) Action(action string, params url.Values) error {
	return task.session.Action(action, params)
}

func (task *Task) SendStat() {
	result := fmt.Sprintf("<b>Task #%d [ %s ]</b>\n\n", task.ID, task.Provider)
	result += fmt.Sprintf("<i>Status:</i> %s\n", webui.StatusToHumanReadable(task.session.Status()))
	stat := task.session.GetStat()
	for title, value := range stat {
		result += fmt.Sprintf("<i>%s:</i> %d\n", title, value)
	}

	msg := bot.Message{Content: result}
	bot.Post(&msg)
}
