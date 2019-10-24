package tinder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/storage"

	"racoondev.tk/gitea/racoon/venera/internal/bot"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/webui"
)

type tinderStat struct {
	Retrieved  uint32
	Liked      uint32
	Superliked uint32
	Passed     uint32
	Errors     uint32
}

type tinderSessionState struct {
	Search           searchSettings
	Top              []ListItem
	Stat             tinderStat
	LastSuperlikeUpd time.Time
}

type tinderSearchSession struct {
	// Защищено мьютексом
	state     tinderSessionState
	status    types.SessionStatus
	lastError error
	mutex     sync.Mutex

	provider TinderProvider
	taskID   uint
	api      *tindergo.TinderGo
	log      *logging.Logger
	rater    *tinderRater
	top      *topList
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

	// TODO: wtf ??
}

func (session *tinderSearchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *tinderSearchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.lastError
}

func (session *tinderSearchSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Tinder session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *tinderSearchSession) Poll() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	now := time.Now()
	if now.Sub(session.state.LastSuperlikeUpd) >= superlikeRefreshPeriod {
		session.state.LastSuperlikeUpd = now
		top := session.top.Get()
		for _, item := range top {
			person, err := storage.LoadPerson(item.ID)
			if err != nil {
				continue
			}

			session.log.Infof("Person '%s' [id = %d, rating = %d] is on top today",
				person.Person.Name, item.ID, item.Rating)

			session.postPerson(person)
		}

		session.top.Clear()
	}
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

	atomic.AddUint32(&session.state.Stat.Superliked, 1)

	return nil
}

func (session *tinderSearchSession) GetStat() map[string]uint32 {
	stat := make(map[string]uint32)
	stat["Retrieved"] = atomic.SwapUint32(&session.state.Stat.Retrieved, 0)
	stat["Liked"] = atomic.SwapUint32(&session.state.Stat.Liked, 0)
	stat["Passed"] = atomic.SwapUint32(&session.state.Stat.Passed, 0)
	stat["Superliked"] = atomic.SwapUint32(&session.state.Stat.Superliked, 0)
	stat["Errors"] = atomic.SwapUint32(&session.state.Stat.Errors, 0)

	return stat
}

func (session *tinderSearchSession) postPerson(person *types.PersonRecord) {
	msg := bot.Message{}
	msg.Content = webui.DecorPerson(&person.Person)
	msg.Photo = person.Person.Photo[0]
	msg.PhotoCaption = person.Person.Name

	actions := session.provider.GetResultActions(person)

	dropAction := types.Action{Title: "Drop",
		Command: fmt.Sprintf("/drop %d", person.ID)}
	actions = append(actions, dropAction)

	if !person.Favourite {
		favouriteAction := types.Action{Title: "Add to Favourites",
			Command: fmt.Sprintf("/favour %d", person.ID)}
		actions = append(actions, favouriteAction)
	}

	msg.Actions = actions

	bot.Post(&msg)
}
