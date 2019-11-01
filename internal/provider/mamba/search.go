package mamba

import (
	"context"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const (
	delayBatchMin = 3 * time.Minute
	delayBatchMax = 5 * time.Minute

	mambaAppID     uint = 2341
	mambaSecretKey      = "3Y3vnn573vt2S4tl6lW8"
)

func (session *mambaSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Mamba API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.api = newMambaRequester(mambaAppID, mambaSecretKey)
	session.mutex.Unlock()

	for {
		var users []mambaUser
		err := session.repeat(ctx, func() error {
			var err error
			users, err = session.api.Search(session.state.Search.AgeFrom,
				session.state.Search.AgeTo,
				session.state.Search.CityID,
				session.state.Offset)
			return err
		})

		if err != nil {
			continue
		}

		if len(users) == 0 {
			session.log.Infof("Mamba session done. Offset = %d", session.state.Offset)
			break
		}

		for _, user := range users {
			session.processUser(ctx, &user)
		}

		session.mutex.Lock()
		session.state.Offset += len(users)
		session.mutex.Unlock()
	}
}

func (session *mambaSearchSession) processUser(ctx context.Context, user *mambaUser) {

	var photos []string
	session.repeat(ctx, func() error {
		var err error
		photos, err = session.api.GetPhotos(user.Info.Oid)
		return err
	})

	var visitTime []time.Time
	err := session.repeat(ctx, func() error {
		var err error
		visitTime, err = session.api.GetLastVisitTime([]int{user.Info.Oid})
		return err
	})

	if err != nil {
		return
	}

	session.log.Debugf("Person '%s' [oid = %d] fetched, photos = %d, visited = %s", user.Info.Name, user.Info.Oid,
		len(photos), visitTime[0].Format("2006-01-02T15:04:05-0700"))
}
