package vk

import "encoding/json"

type vkSessionState struct {
	Search searchSettings
}

type vkSearchSession struct {
	state vkSessionState
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

func NewSession(search *searchSettings) *vkSearchSession {
	return &vkSearchSession{state: vkSessionState{Search: *search}}
}
