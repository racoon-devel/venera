package tinder

import "encoding/json"

type tinderSessionState struct {
	Search searchSettings
}

type tinderSearchSession struct {
	state tinderSessionState
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

func NewSession(search *searchSettings) *tinderSearchSession {
	return &tinderSearchSession{state: tinderSessionState{Search: *search}}
}
