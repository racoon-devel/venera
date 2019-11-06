package export

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"sync/atomic"


	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

type exportStat struct {
	Retrieved uint32
	Errors    uint32
	Progress  uint32
	Processed uint32
}

type exportState struct {
	Search   searchSettings
	Stat     exportStat
	Offset   uint
	FileName string
}

type exportSession struct {
	// Защищено мьютексом
	state     exportState
	status    types.SessionStatus
	lastError error
	mutex     sync.Mutex

	provider ExportProvider
	taskID   uint
	log      *logging.Logger

	total uint
	tw    *tar.Writer
}

func (session *exportSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *exportSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return json.Unmarshal([]byte(state), &session.state)
}

func (session *exportSession) Reset() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	atomic.StoreUint32(&session.state.Stat.Errors, 0)
	atomic.StoreUint32(&session.state.Stat.Retrieved, 0)
	atomic.StoreUint32(&session.state.Stat.Processed, 0)
	atomic.StoreUint32(&session.state.Stat.Progress, 0)

	session.state.Offset = 0
	session.total = 0

	os.Remove(session.state.FileName)
}

func (session *exportSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *exportSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
}

func (session *exportSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Export session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *exportSession) Poll() {
}

func (session *exportSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	var ctx struct {
		Stat exportStat
		FileName string
		Url template.URL
		Ready bool
	}

	ctx.Stat = session.state.Stat
	ctx.Stat.Processed = ctx.Stat.Processed / (1024 * 1024)

	_, ctx.FileName = path.Split(session.state.FileName)
	ctx.Url = template.URL(fmt.Sprintf("/export/download/%s", ctx.FileName))

	session.mutex.Lock()
	ctx.Ready = session.status == types.StatusDone
	session.mutex.Unlock()

	webui.DisplayEditTask(w, "export", &ctx)

	return false, nil
}

func (session *exportSession) Action(action string, params url.Values) error {
	return fmt.Errorf("export: undefined action: '%s'", action)
}

func (session *exportSession) GetStat() map[string]uint32 {
	stat := make(map[string]uint32)
	stat["Retrieved"] = atomic.SwapUint32(&session.state.Stat.Retrieved, 0)
	stat["Errors"] = atomic.SwapUint32(&session.state.Stat.Errors, 0)
	stat["Processed"] = atomic.SwapUint32(&session.state.Stat.Processed, 0)
	stat["Progress"] = atomic.LoadUint32(&session.state.Stat.Progress)

	return stat
}

func (session *exportSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("export: %+v", err)
	session.status = types.StatusError
	session.lastError = err
	atomic.AddUint32(&session.state.Stat.Errors, 1)
}
