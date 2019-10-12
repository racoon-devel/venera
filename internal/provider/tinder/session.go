package tinder

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
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

	api   *tindergo.TinderGo
	log   *logging.Logger
	rater *tinderRater
	top   *topList
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

func NewSession(search *searchSettings, log *logging.Logger) *tinderSearchSession {
	return &tinderSearchSession{state: tinderSessionState{Search: *search}, log: log}
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
