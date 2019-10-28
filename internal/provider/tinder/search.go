package tinder

import (
	"context"
	"math/rand"
	"racoondev.tk/gitea/racoon/venera/internal/bot"
	"sync/atomic"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/storage"

	"racoondev.tk/gitea/racoon/venera/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	maxSuperLikes          uint = 5
	superlikeRefreshPeriod      = 140 * time.Second

	delayBatchMin     = 3 * time.Minute
	delayBatchMax     = 5 * time.Minute
	requestPerSession = 3

	delaySessionMin = 1 * time.Hour
	delaySessionMax = 3 * time.Hour
)

func (session *tinderSearchSession) auth(ctx context.Context) error {
	for {
		auth := newTinderAuth(session.state.Search.Tel)
		if err := auth.RequestCode(); err != nil {
			return err
		}

		code, err := bot.Request(ctx, "Require Tinder authentification token")
		if err != nil {
			utils.Delay(ctx, utils.Range{Min: delayBatchMin, Max: delayBatchMax})
			continue
		}

		auth.LoginCode = code
		if err := auth.RequestToken(); err != nil {
			session.raise(err)
			utils.Delay(ctx, utils.Range{Min: delayBatchMin, Max: delayBatchMax})
			continue
		}

		session.log.Infof("Tinder API token retrieved: %s", auth.APIToken)

		session.mutex.Lock()
		session.state.Search.APIToken = auth.APIToken
		session.api.SetAPIToken(auth.APIToken)
		session.mutex.Unlock()

		return nil
	}
}

func (session *tinderSearchSession) setup(ctx context.Context) {
	session.log.Debugf("tinder: check auth")

	profile, err := session.api.Profile()
	if err != nil || profile.ID == "" {
		if err := session.auth(ctx); err != nil {
			session.raise(err)
			return
		}
	}

	session.log.Info(profile)

	session.log.Debugf("tinder: update location...")

	session.repeat(ctx, func() error {
		return session.api.UpdateLocation(session.state.Search.Latitude, session.state.Search.Longitude)
	})

	session.log.Debugf("tinder: update preferences...")

	session.mutex.Lock()
	pref := tindergo.SearchPreferences{
		AgeFilterMin:   int(session.state.Search.AgeFrom),
		AgeFilterMax:   int(session.state.Search.AgeTo),
		DistanceFilter: 10,
		GenderFilter:   1,
	}
	session.mutex.Unlock()

	session.repeat(ctx, func() error {
		return session.api.UpdateSearchPreferences(pref)
	})
}

func (session *tinderSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Tinder API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning

	session.api = tindergo.New()
	session.api.SetAPIToken(session.state.Search.APIToken)

	session.rater = &tinderRater{}
	session.rater.Init(session.log, &session.state.Search.SearchSettings)

	if session.top == nil {
		session.top = newTopList(maxSuperLikes)
	}

	if session.state.Matches == nil {
		session.state.Matches = make(map[string]types.Person)
	}

	if session.state.LastSuperlikeUpd.IsZero() {
		session.state.LastSuperlikeUpd = time.Now()
	}

	session.mutex.Unlock()

	session.setup(ctx)

	for {
		for i := 0; i < requestPerSession; i++ {
			session.processBatch(ctx)
			session.log.Info("tinder: processing batch finished")
			utils.Delay(ctx, utils.Range{Min: delayBatchMin, Max: delayBatchMax})
		}

		session.log.Info("tinder: processing session finished")
		utils.Delay(ctx, utils.Range{Min: delaySessionMin, Max: delaySessionMax})

		session.repeat(ctx, func() error {
			return session.api.UpdateLocation(
				utils.RandomCoordinate(session.state.Search.Latitude),
				utils.RandomCoordinate(session.state.Search.Longitude),
			)
		})
	}
}

func (session *tinderSearchSession) processBatch(ctx context.Context) {

	var persons []tindergo.RecsCoreUser
	err := session.repeat(ctx, func() error {
		var err error
		persons, err = session.api.RecsCore()
		return err
	})

	if err != nil {
		session.log.Errorf("Retrieve persons failed: %+v", err)
		session.auth(ctx)
	}

	session.log.Debugf("tinder: got %d persons", len(persons))

	for _, record := range persons {
		atomic.AddUint32(&session.state.Stat.Retrieved, 1)
		session.log.Debugf("Rate person '%s'...", record.Name)
		person := convertPersonRecord(&record)
		rating := session.rater.Rate(&person)

		if person.Bio != "" {
			session.log.Debug(person.Bio)
		}

		toLike := rand.Intn(2)

		if rating > 0 || (rating == 0 && toLike == 1) {
			session.log.Debugf("Like '%s' rating(%d)", person.Name, rating)
			session.repeat(ctx, func() error {
				_, err := session.api.Like(record)
				return err
			})
			atomic.AddUint32(&session.state.Stat.Liked, 1)

			if rating > 0 {
				session.appendResult(&person)
			}

		} else {
			session.log.Debugf("Dislike '%s' rating(%d)", person.Name, 0)
			session.repeat(ctx, func() error {
				_, err := session.api.Pass(record)
				return err
			})
			atomic.AddUint32(&session.state.Stat.Passed, 1)
		}
	}
}

func (session *tinderSearchSession) appendResult(person *types.Person) {
	id, err := storage.AppendPerson(person, session.taskID, session.provider.ID())
	if err == nil {
		session.mutex.Lock()
		session.top.Push(id, person.Rating)
		session.mutex.Unlock()
	}
}
