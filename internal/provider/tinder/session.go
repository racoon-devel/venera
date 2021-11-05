package tinder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ccding/go-logging/logging"

	tindergo "github.com/racoon-devel/TinderGo"
	"github.com/racoon-devel/venera/internal/bot"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
	"github.com/racoon-devel/venera/internal/webui"
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
	Matches          map[string]types.Person
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
	rater    types.Rater
	top      *topList
	geo      geoposition
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

	session.state.LastSuperlikeUpd = time.Time{}
	atomic.StoreUint32(&session.state.Stat.Errors, 0)
	atomic.StoreUint32(&session.state.Stat.Liked, 0)
	atomic.StoreUint32(&session.state.Stat.Passed, 0)
	atomic.StoreUint32(&session.state.Stat.Retrieved, 0)
	atomic.StoreUint32(&session.state.Stat.Superliked, 0)

	session.state.Top = make([]ListItem, 0)
}

func (session *tinderSearchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *tinderSearchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
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

	matches, err := session.api.Matches()
	if err == nil {
		for _, match := range matches {
			_, isUnique := session.state.Matches[match.ID]
			if !isUnique {
				session.log.Infof("New match found! %s { isSuperliked: %t, isBoosted:%t }", match.Person.Name,
					match.IsSuperLike, match.IsBoostMatch)
				person := convertMatch(&match)
				session.state.Matches[match.ID] = person
				session.postMatchPerson(match.ID, &person)
			}
		}
	}

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
		session.state.Search.Longitude = search.Longitude
		session.state.Search.Latitude = search.Latitude

		return true, nil
	}

	type editContext struct {
		URL       string
		Likes     string
		Dislikes  string
		AgeFrom   uint
		AgeTo     uint
		Latitude  float32
		Longitude float32
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	ctx := editContext{URL: r.URL.String(),
		AgeFrom:   session.state.Search.AgeFrom,
		AgeTo:     session.state.Search.AgeTo,
		Latitude:  session.state.Search.Latitude,
		Longitude: session.state.Search.Longitude,
	}

	ctx.Likes = utils.ListToString(session.state.Search.Likes)
	ctx.Dislikes = utils.ListToString(session.state.Search.Dislikes)

	session.log.Debugf("Display edit page")

	webui.DisplayEditTask(w, "tinder", &ctx)

	return false, nil
}

func (session *tinderSearchSession) superlike(userID string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	api := tindergo.New()
	api.SetAPIToken(session.state.Search.APIToken)

	session.log.Debugf("Superlike %s...", userID)

	resp, err := api.SuperLike(userID, "")
	if err != nil {
		return err
	}

	session.log.Debug(resp)

	if resp.LimitExceeded {
		return fmt.Errorf("Superlike limit exceeded")
	}

	if resp.Status != 200 {
		return fmt.Errorf("Superlike failed: %d", resp.Status)
	}

	atomic.AddUint32(&session.state.Stat.Superliked, 1)

	return nil
}

func (session *tinderSearchSession) unmatch(matchID string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	api := tindergo.New()
	api.SetAPIToken(session.state.Search.APIToken)

	err := api.Unmatch(matchID)
	if err != nil {
		return err
	}

	delete(session.state.Matches, matchID)

	session.log.Infof("Match '%s' unmatched", matchID)

	return nil
}

func (session *tinderSearchSession) Action(action string, params url.Values) error {
	IDs, ok := params["id"]
	if !ok || len(IDs) == 0 {
		return fmt.Errorf("user ID missed")
	}

	ID := IDs[0]

	switch action {
	case "superlike":
		return session.superlike(ID)

	case "unmatch":
		return session.unmatch(ID)

	default:
		return fmt.Errorf("tinder: undefined action: '%s'", action)
	}
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

func (session *tinderSearchSession) postMatchPerson(matchID string, person *types.Person) {
	msg := bot.Message{}
	msg.Content = webui.DecorPerson(person)

	if len(person.Photo) != 0 {
		msg.Photo = person.Photo[0]
		msg.PhotoCaption = person.Name
	}

	msg.Actions = []types.Action{
		types.Action{Title: "Unmatch", Command: fmt.Sprintf("/action %d unmatch id %s",
			session.taskID, matchID)},
	}

	bot.Post(&msg)
}
