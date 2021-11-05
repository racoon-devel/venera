package mamba

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/racoon-devel/venera/internal/utils"

	"github.com/ccding/go-logging/logging"

	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/webui"
)

type mambaStat struct {
	Retrieved uint32
	Liked     uint32
	Disliked  uint32
	Errors    uint32
}

type mambaSessionState struct {
	Search searchSettings
	Stat   mambaStat
	Offset int
}

type mambaSearchSession struct {
	// Защищено мьютексом
	state     mambaSessionState
	status    types.SessionStatus
	lastError error
	mutex     sync.Mutex

	provider   MambaProvider
	taskID     uint
	api        *mambaRequester
	log        *logging.Logger
	rater      types.Rater
	lookForExp *regexp.Regexp
}

func (session *mambaSearchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *mambaSearchSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return json.Unmarshal([]byte(state), &session.state)
}

func (session *mambaSearchSession) Reset() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	atomic.StoreUint32(&session.state.Stat.Errors, 0)
	atomic.StoreUint32(&session.state.Stat.Retrieved, 0)
	atomic.StoreUint32(&session.state.Stat.Liked, 0)
	atomic.StoreUint32(&session.state.Stat.Disliked, 0)
	session.state.Offset = 0
}

func (session *mambaSearchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *mambaSearchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
}

func (session *mambaSearchSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Mamba session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *mambaSearchSession) Poll() {
}

func (session *mambaSearchSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	if r.Method == "POST" {

		search, err := parseForm(r, true)
		if err != nil {
			return false, err
		}

		api := newMambaRequester(mambaAppID, mambaSecretKey)
		cityID, err := api.GetCityID(search.City)
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
		session.state.Offset = 0
		session.state.Search.CityID = cityID

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

	webui.DisplayEditTask(w, "mamba", &ctx)

	return false, nil
}

func (session *mambaSearchSession) Action(action string, params url.Values) error {
	IDs, ok := params["id"]
	if !ok || len(IDs) == 0 {
		return fmt.Errorf("user ID missed")
	}

	ID := IDs[0]

	switch action {
	case "open":
		return session.open(ID)

	default:
		return fmt.Errorf("mamba: undefined action: '%s'", action)
	}
}

func (session *mambaSearchSession) GetStat() map[string]uint32 {
	stat := make(map[string]uint32)
	stat["Retrieved"] = atomic.SwapUint32(&session.state.Stat.Retrieved, 0)
	stat["Errors"] = atomic.SwapUint32(&session.state.Stat.Errors, 0)
	stat["Liked"] = atomic.SwapUint32(&session.state.Stat.Liked, 0)
	stat["Disliked"] = atomic.SwapUint32(&session.state.Stat.Disliked, 0)

	return stat
}

func (session *mambaSearchSession) open(ID string) error {
	// TODO
	return nil
}
