package tinder

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"

	tindergo "github.com/racoon-devel/TinderGo"
)

const (
	maxSuperLikes          uint = 5
	superlikeRefreshPeriod      = 24 * time.Hour

	delayBatchMin     = 3 * time.Minute
	delayBatchMax     = 5 * time.Minute
	requestPerSession = 7

	delaySessionMin = 1 * time.Hour
	delaySessionMax = 2 * time.Hour

	delayDecideMin = 3 * time.Second
	delayDecideMax = 20 * time.Second

	apiTokenRefreshInterval = 2 * 24 * time.Hour
)

func (session *tinderSearchSession) auth(ctx context.Context) error {
	auth := newTinderAuth(session.state.Search.Login, session.state.Search.Password)
	err := auth.SignIn(session.log)
	if err == nil {
		session.mutex.Lock()
		session.state.Search.APIToken = auth.APIToken
		session.state.LastAuthTime = time.Now()
		session.api.SetAPIToken(auth.APIToken)
		session.mutex.Unlock()
		return nil
	} else {
		session.log.Warnf("Login failed: %+v", err)
	}

	return err
}

func (session *tinderSearchSession) expired() bool {
	return session.state.Search.APIToken == "" || time.Since(session.state.LastAuthTime) >= apiTokenRefreshInterval
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

	center := geoposition{session.state.Search.Latitude, session.state.Search.Longitude}
	session.geo = randomPosition(center)

	session.repeat(ctx, func() error {
		return session.api.UpdateLocation(session.geo.Latitude, session.geo.Longitude)
	})

	session.log.Debugf("tinder: update preferences...")

	session.mutex.Lock()
	pref := tindergo.SearchPreferences{
		AgeFilterMin:   int(session.state.Search.AgeFrom),
		AgeFilterMax:   int(session.state.Search.AgeTo),
		DistanceFilter: int(session.state.Search.Distance),
		GenderFilter:   1,
	}
	session.mutex.Unlock()

	session.repeat(ctx, func() error {
		return session.api.UpdateSearchPreferences(pref)
	})
}

func (session *tinderSearchSession) retrySignIn(ctx context.Context) {
	if session.expired() {
		err := session.repeat(ctx, func() error {
			return session.auth(ctx)
		})
		if err != nil {
			session.raise(err)
			return
		}
	}
}

func (session *tinderSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Tinder API Session....")

	session.api = tindergo.New()
	session.api.SetAPIToken(session.state.Search.APIToken)

	session.retrySignIn(ctx)

	session.mutex.Lock()
	session.status = types.StatusRunning

	session.rater = rater.NewRater(session.state.Search.Rater, "tinder", session.log, &session.state.Search.SearchSettings)
	defer session.rater.Close()

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
		session.retrySignIn(ctx)
		for i := 0; i < requestPerSession; i++ {
			session.processBatch(ctx)
			session.log.Info("tinder: processing batch finished")
			utils.Delay(ctx, utils.Range{Min: delayBatchMin, Max: delayBatchMax})
		}

		session.log.Info("tinder: processing session finished")
		utils.Delay(ctx, utils.Range{Min: delaySessionMin, Max: delaySessionMax})

		// Пока отключил "эмулятор ходьбы по городу"

		//center := geoposition{session.state.Search.Latitude, session.state.Search.Longitude}
		//shiftPosition(center, &session.geo)

		//session.repeat(ctx, func() error {
		//	return session.api.UpdateLocation(
		//		session.geo.Latitude,
		//		session.geo.Longitude,
		//	)
		//})
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
		session.log.Debugf("Ping time: %s", record.PingTime.Format("Mon Jan _2 15:04:05 2006"))
		person := convertPersonRecord(&record)

		var rating int
		stored := storage.SearchPerson(session.provider.ID(), person.UserID)
		if stored != nil {
			rating = stored.Rating
		} else {
			rating = session.rater.Rate(&person)
		}

		toLike := rand.Intn(2)

		if rating >= session.rater.Threshold(types.LikeThreshold) || (rating > 0 && toLike == 1) {
			session.log.Debugf("Like '%s' rating(%d)", person.Name, rating)
			session.repeat(ctx, func() error {
				_, err := session.api.Like(record)
				return err
			})
			atomic.AddUint32(&session.state.Stat.Liked, 1)

			if rating > 0 && stored == nil {
				session.appendResult(&person)
			}

		} else {
			session.log.Debugf("Dislike '%s' rating(%d)", person.Name, 0)
			session.repeat(ctx, func() error {
				_, err := session.api.Pass(record)
				return err
			})
			atomic.AddUint32(&session.state.Stat.Passed, 1)

			// все равно сохраним в базу для последующего rerate
			if stored == nil {
				_, _ = storage.AppendPerson(&person, session.taskID, session.provider.ID())
			}
		}

		utils.Delay(ctx, utils.Range{Min: delayDecideMin, Max: delayDecideMax})
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
