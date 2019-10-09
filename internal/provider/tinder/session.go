package tinder

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ccding/go-logging/logging"

	"github.com/DiSiqueira/TinderGo"
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

func (session *tinderSearchSession) Process(ctx *context.Context) {
	session.log.Debugf("Starting Tinder API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.mutex.Unlock()

	session.api = tindergo.New()
	err := session.api.Authenticate(session.state.Search.Token)
	if err != nil {
		session.log.Errorf("tinder: authenticate failed: %+v", err)
	}
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
