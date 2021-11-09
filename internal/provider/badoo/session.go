package badoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/racoon-devel/venera/internal/provider/badoo/badoogo"

	"github.com/racoon-devel/venera/internal/utils"

	"github.com/ccding/go-logging/logging"

	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/webui"
)

type badooStat struct {
	Retrieved uint32
	Liked     uint32
	Disliked  uint32
	Errors    uint32
}

type badooSessionState struct {
	Search searchSettings
	Stat   badooStat
}

type badooSearchSession struct {
	// Защищено мьютексом
	state     badooSessionState
	status    types.SessionStatus
	lastError error
	mutex     sync.Mutex

	provider BadooProvider
	taskID   uint
	log      *logging.Logger
	browser  *badoogo.BadooRequester
	liker    *badoogo.BadooRequester
	walker   *badoogo.BadooRequester
	rater    types.Rater

	alcoExpr  *regexp.Regexp
	smokeExpr *regexp.Regexp
	bodyExpr  *regexp.Regexp

	errorCounter int
}

func (session *badooSearchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *badooSearchSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	return json.Unmarshal([]byte(state), &session.state)
}

func (session *badooSearchSession) Reset() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	atomic.StoreUint32(&session.state.Stat.Errors, 0)
	atomic.StoreUint32(&session.state.Stat.Retrieved, 0)
	atomic.StoreUint32(&session.state.Stat.Liked, 0)
	atomic.StoreUint32(&session.state.Stat.Disliked, 0)
}

func (session *badooSearchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *badooSearchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
}

func (session *badooSearchSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Badoo session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *badooSearchSession) Poll() {
}

func (session *badooSearchSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	if r.Method == "POST" {
		search, err := parseForm(r, true)

		if err != nil {
			return false, err
		}

		session.mutex.Lock()
		defer session.mutex.Unlock()

		session.state.Search.Likes = search.Likes
		session.state.Search.Dislikes = search.Dislikes
		session.state.Search.Longitude = search.Longitude
		session.state.Search.Latitude = search.Latitude

		return true, nil
	}

	type editContext struct {
		URL       string
		Latitude  float32
		Longitude float32
		Likes     string
		Dislikes  string
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	ctx := editContext{URL: r.URL.String(),
		Latitude:  session.state.Search.Latitude,
		Longitude: session.state.Search.Longitude,
	}

	ctx.Likes = utils.ListToString(session.state.Search.Likes)
	ctx.Dislikes = utils.ListToString(session.state.Search.Dislikes)

	session.log.Debugf("Display edit page")

	webui.DisplayEditTask(w, "badoo", &ctx)

	return false, nil
}

func (session *badooSearchSession) Action(action string, params url.Values) error {
	return fmt.Errorf("badoo: undefined action: '%s'", action)
}

func (session *badooSearchSession) GetStat() map[string]uint32 {
	stat := make(map[string]uint32)
	stat["Retrieved"] = atomic.SwapUint32(&session.state.Stat.Retrieved, 0)
	stat["Errors"] = atomic.SwapUint32(&session.state.Stat.Errors, 0)
	stat["Liked"] = atomic.SwapUint32(&session.state.Stat.Liked, 0)
	stat["Disliked"] = atomic.SwapUint32(&session.state.Stat.Disliked, 0)

	return stat
}
