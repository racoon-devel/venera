package tinder

import (
	"encoding/json"

	"github.com/ccding/go-logging/logging"

	"github.com/DiSiqueira/TinderGo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type tinderSessionState struct {
	Search searchSettings
}

type tinderSearchSession struct {
	state  tinderSessionState
	api    *tindergo.TinderGo
	status types.SessionStatus
	log    *logging.Logger
}

func (ctx *tinderSearchSession) SaveState() string {
	data, err := json.Marshal(&ctx.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (ctx *tinderSearchSession) LoadState(state string) error {
	err := json.Unmarshal([]byte(state), &ctx.state)
	return err
}

func (ctx *tinderSearchSession) process() {
	ctx.log.Debugf("Starting Tinder API Session....")
	ctx.api = tindergo.New()
	err := ctx.api.Authenticate(ctx.state.Search.Token)
	if err != nil {
		ctx.log.Errorf("tinder: authenticate failed: %+v", err)
	}
}

func (ctx *tinderSearchSession) Start() {
	ctx.status = types.StatusRunning
	go ctx.process()
}

func (ctx *tinderSearchSession) Stop() {

}

func (ctx *tinderSearchSession) Reset() {

}

func (ctx *tinderSearchSession) Status() types.SessionStatus {
	return ctx.status
}

func NewSession(search *searchSettings, log *logging.Logger) *tinderSearchSession {
	return &tinderSearchSession{state: tinderSessionState{Search: *search}}
}
