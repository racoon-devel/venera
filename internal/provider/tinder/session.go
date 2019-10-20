package tinder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

type tinderSessionState struct {
	Search searchSettings
	Top    []types.Person
}

type tinderSearchSession struct {
	// Защищено мьютексом
	state     tinderSessionState
	status    types.SessionStatus
	lastError error
	mutex     sync.Mutex
	results   []*types.Person

	api   *tindergo.TinderGo
	log   *logging.Logger
	rater *tinderRater
	top   *topList
}

func NewSession(search *searchSettings, log *logging.Logger) *tinderSearchSession {
	return &tinderSearchSession{state: tinderSessionState{Search: *search}, log: log}
}

func (session *tinderSearchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.top != nil {
		session.state.Top = session.top.Get()
	}

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *tinderSearchSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	err := json.Unmarshal([]byte(state), &session.state)
	if err == nil {
		session.top = loadTopList(maxSuperLikes, session.state.Top)
	}
	fmt.Println(err)
	return err
}

func (session *tinderSearchSession) Reset() {
	session.mutex.Lock()
	defer session.mutex.Unlock()
}

func (session *tinderSearchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *tinderSearchSession) Process(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Tinder session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()
	session.process(ctx)
}

func (session *tinderSearchSession) Results() []*types.Person {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	result := session.results
	session.results = make([]*types.Person, 0)
	return result
}

func listToString(list []string) string {
	var result string
	for _, item := range list {
		result += item + ","
	}

	if len(result) != 0 {
		result = strings.TrimSuffix(result, result[len(result)-1:])
	}

	return result
}

func (session *tinderSearchSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	if r.Method == "POST" {

		search, _, err := parseForm(r, true)
		if err != nil {
			return false, err
		}

		session.mutex.Lock()
		defer session.mutex.Unlock()

		session.state.Search.AgeFrom = search.AgeFrom
		session.state.Search.AgeTo = search.AgeTo
		session.state.Search.Likes = search.Likes
		session.state.Search.Dislikes = search.Dislikes

		return true, nil
	}

	type editContext struct {
		URL      string
		Likes    string
		Dislikes string
		AgeFrom  uint
		AgeTo    uint
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	ctx := editContext{URL: r.URL.String(), AgeFrom: session.state.Search.AgeFrom, AgeTo: session.state.Search.AgeTo}
	ctx.Likes = listToString(session.state.Search.Likes)
	ctx.Dislikes = listToString(session.state.Search.Dislikes)

	session.log.Debugf("Display edit page")

	webui.DisplayEditTask(w, "tinder", &ctx)

	return false, nil
}

func (session *tinderSearchSession) Action(action string, params url.Values) error {
	if action != "superlike" {
		return fmt.Errorf("tinder: undefined action: '%s'", action)
	}

	IDs, ok := params["id"]
	if !ok || len(IDs) == 0 {
		return fmt.Errorf("user ID missed")
	}

	ID := IDs[0]

	session.mutex.Lock()
	defer session.mutex.Unlock()

	api := tindergo.New()
	api.SetAPIToken(session.state.Search.APIToken)

	resp, err := api.SuperLike(ID, "")
	if err != nil {
		return err
	}

	if resp.LimitExceeded {
		return fmt.Errorf("Superlike limit exceeded")
	}

	if resp.Status != 200 {
		return fmt.Errorf("Superlike failed: %d", resp.Status)
	}

	return nil

}
