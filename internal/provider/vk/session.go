package vk

import (
	"context"
	"encoding/json"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type vkSessionState struct {
	Search searchSettings
}

type vkSearchSession struct {
	state  vkSessionState
	status types.SessionStatus
	done   chan bool
}

func newSession() *vkSearchSession {
	return &vkSearchSession{
		done: make(chan bool),
	}
}

func createSession(search *searchSettings) *vkSearchSession {
	session := newSession()
	session.state = vkSessionState{Search: *search}
	return session
}

func (ctx *vkSearchSession) SaveState() string {
	data, err := json.Marshal(&ctx.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (ctx *vkSearchSession) LoadState(state string) error {
	err := json.Unmarshal([]byte(state), &ctx.state)
	return err
}

func (session *vkSearchSession) Process(ctx context.Context) {
	session.status = types.StatusRunning
	//go ctx.do()
}

func (ctx *vkSearchSession) Reset() {

}

func (ctx *vkSearchSession) Status() types.SessionStatus {
	return ctx.status
}
