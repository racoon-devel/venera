package vk

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/types"
	"time"
)

func (session *searchSession) checkAuth() error {
	if session.state.AccessToken == "" || time.Since(session.state.LastAuthTime) >= 24*time.Hour {
		return session.signIn()
	}

	return nil
}

func (session *searchSession) process(ctx context.Context) {
	session.log.Debugf("VK starting session...")
	if err := session.checkAuth(); err != nil {
		session.raise(err)
		return
	}

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.api = api.NewVK(session.state.AccessToken)
	session.rater = rater.NewRater(session.state.Search.Rater, "vk", session.log, &session.state.Search.SearchSettings)
	defer session.rater.Close()
	session.mutex.Unlock()

	<-ctx.Done()
}
