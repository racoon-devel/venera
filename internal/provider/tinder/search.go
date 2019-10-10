package tinder

import (
	"context"

	"github.com/DiSiqueira/TinderGo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func (session *tinderSearchSession) Process(ctx context.Context) {
	session.log.Debugf("Starting Tinder API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.mutex.Unlock()

	session.api = tindergo.New()
	err := session.api.Authenticate(session.state.Search.Token)
	if err != nil {
		session.log.Errorf("tinder: authenticate failed: %+v", err)
	}

	req := tindergo.NewRequest()
	req.Get("")
	//profile, err :=  session.api.Profile()

}
