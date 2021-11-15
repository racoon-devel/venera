package vk

import (
	"context"
	"encoding/json"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
	"github.com/racoon-devel/venera/internal/webui"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type searchSession struct {
	// Защищено мьютексом
	status    types.SessionStatus
	lastError error
	state     sessionState
	mutex     sync.Mutex

	provider Provider
	taskID   uint
	log      *logging.Logger
	rater    types.Rater

	api *api.VK
	ctx context.Context
}

func (session *searchSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Vk session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *searchSession) Reset() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.state.State = stateInitialize
	session.state.UserSearch = userSearch{}
	session.state.GroupSearch = groupSearch{}
	session.state.InGroupSearch = inGroupSearch{}
	session.state.LastAuthTime = time.Time{}

	atomic.StoreUint32(&session.state.Stat.Errors, 0)
	atomic.StoreUint32(&session.state.Stat.Saved, 0)
	atomic.StoreUint32(&session.state.Stat.Retrieved, 0)
	atomic.StoreUint32(&session.state.Stat.Groups, 0)
}

func (session *searchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *searchSession) GetStat() map[string]uint32 {
	stat := make(map[string]uint32)
	stat["Retrieved"] = atomic.SwapUint32(&session.state.Stat.Retrieved, 0)
	stat["Saved"] = atomic.SwapUint32(&session.state.Stat.Saved, 0)
	stat["Errors"] = atomic.SwapUint32(&session.state.Stat.Errors, 0)
	stat["Groups"] = atomic.SwapUint32(&session.state.Stat.Groups, 0)

	return stat
}

func (session *searchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
}

func (session *searchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *searchSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	err := json.Unmarshal([]byte(state), &session.state)
	return err
}

func (session *searchSession) Poll() {
}

func (session *searchSession) Action(action string, params url.Values) error {
	panic("implement me")
}

func (session *searchSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	if r.Method == "POST" {

		search, err := parseForm(r, true)
		if err != nil {
			return false, err
		}

		session.mutex.Lock()
		defer session.mutex.Unlock()

		session.state.Search.AgeFrom = search.AgeFrom
		session.state.Search.AgeTo = search.AgeTo
		session.state.Search.Likes = search.Likes
		session.state.Search.Dislikes = search.Dislikes
		session.state.Search.City = search.City

		return true, nil
	}

	type editContext struct {
		URL      string
		Likes    string
		Dislikes string
		AgeFrom  uint
		AgeTo    uint
		City     string
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	ctx := editContext{URL: r.URL.String(),
		AgeFrom: session.state.Search.AgeFrom,
		AgeTo:   session.state.Search.AgeTo,
		City:    session.state.Search.City,
	}

	ctx.Likes = utils.ListToString(session.state.Search.Likes)
	ctx.Dislikes = utils.ListToString(session.state.Search.Dislikes)

	session.log.Debugf("Display edit page")

	webui.DisplayEditTask(w, "vk", &ctx)

	return false, nil
}
