package mamba

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/storage"

	"racoondev.tk/gitea/racoon/venera/internal/rater"
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
	session.rater = rater.NewRater("default", session.log, &session.state.Search.SearchSettings)
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

	person := convertPersonRecord(user, photos)
	rating := session.rater.Rate(&person)
	person.Rating = rating
	// TODO: extra rating

	session.log.Debugf("Person '%s' [oid = %d, photos = %d, visited = %s] fetched: %d", user.Info.Name, user.Info.Oid,
		len(photos), visitTime[0].Format("2006-01-02T15:04:05-0700"), rating)

	atomic.AddUint32(&session.state.Stat.Retrieved, 1)

	if rating > 0 {
		atomic.AddUint32(&session.state.Stat.Liked, 1)
		storage.AppendPerson(&person, session.taskID, session.provider.ID())
	} else {
		atomic.AddUint32(&session.state.Stat.Disliked, 1)
	}
}

func convertPersonRecord(record *mambaUser, extraPhotos []string) types.Person {
	person := types.Person{
		UserID: strconv.Itoa(record.Info.Oid),
		Name:   record.Info.Name,
		Bio:    record.About,
	}

	person.Age = uint(record.Info.Age)
	person.Photo = make([]string, 1, len(extraPhotos)+1)
	person.Photo[0] = record.Info.Photo

	if len(extraPhotos) != 0 {
		person.Photo = append(person.Photo, extraPhotos...)
	}

	return person
}
