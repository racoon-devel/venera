package tinder

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type tinderSessionState struct {
	Search searchSettings
}

type tinderSearchSession struct {
	state     tinderSessionState
	api       *tindergo.TinderGo
	status    types.SessionStatus
	log       *logging.Logger
	mutex     sync.Mutex
	lastError error
}

func (session *tinderSearchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

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
