package tinder

import (
	"context"
	"math/rand"

	"racoondev.tk/gitea/racoon/tindergo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	maxSuperLikes uint = 5
	// TODO: в SearchSettings
	locationLatitude   float32 = 55.741676
	locationLongtitude float32 = 37.624928

	delayBatchMinMs   uint32 = 3.0 * 60 * 1000
	delayBatchMaxMs   uint32 = 5.0 * 60 * 1000
	requestPerSession        = 3

	delaySessionMinMs uint32 = 1.0 * 60 * 60 * 1000
	delaySessionMaxMs uint32 = 3.0 * 60 * 60 * 1000
)

func (session *tinderSearchSession) setup(ctx context.Context) {
	session.log.Debugf("tinder: authentification...")

	session.repeat(ctx, func() error {
		return session.api.Authenticate(session.state.Search.Token)
	})

	session.log.Debugf("tinder: update location...")

	session.repeat(ctx, func() error {
		return session.api.UpdateLocation(locationLatitude, locationLongtitude)
	})

	session.log.Debugf("tinder: update preferences...")

	pref := tindergo.SearchPreferences{
		AgeFilterMin:   int(session.state.Search.AgeFrom),
		AgeFilterMax:   int(session.state.Search.AgeTo),
		DistanceFilter: 10,
		GenderFilter:   1,
	}

	session.log.Debug(pref.AgeFilterMin, pref.AgeFilterMax)

	session.repeat(ctx, func() error {
		return session.api.UpdateSearchPreferences(pref)
	})
}

func (session *tinderSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Tinder API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.results = make([]*types.Person, 0)
	session.mutex.Unlock()

	session.api = tindergo.New()
	session.rater = &tinderRater{}
	session.rater.Init(session.log, &session.state.Search.SearchSettings)

	if session.top == nil {
		session.top = newTopList(maxSuperLikes)
	}

	session.setup(ctx)

	for {
		for i := 0; i < requestPerSession; i++ {
			session.processBatch(ctx)
			session.log.Info("tinder: processing batch finished")
			utils.Delay(ctx, utils.Range{MinMs: delayBatchMinMs, MaxMs: delayBatchMaxMs})
		}

		session.log.Info("tinder: processing session finished")
		utils.Delay(ctx, utils.Range{MinMs: delaySessionMinMs, MaxMs: delaySessionMaxMs})
	}
}

func (session *tinderSearchSession) processBatch(ctx context.Context) {

	var persons []tindergo.RecsCoreUser
	session.repeat(ctx, func() error {
		var err error
		persons, err = session.api.RecsCore()
		return err
	})

	session.log.Debugf("tinder: got %d persons", len(persons))

	for _, record := range persons {
		session.log.Debugf("Rate person '%s'...", record.Name)
		person := convert(&record)
		rating := session.rater.Rate(&person)

		if person.Bio != "" {
			session.log.Debug(person.Bio)
		}

		toLike := rand.Intn(2)

		if rating > 0 || (rating == 0 && toLike == 1) {
			session.top.Push(person)
			session.log.Debugf("Like '%s' rating(%d)", person.Name, rating)
			session.repeat(ctx, func() error {
				_, err := session.api.Like(record)
				return err
			})

			if rating > 0 {
				session.appendResult(&person)
			}

		} else {
			session.log.Debugf("Dislike '%s' rating(%d)", person.Name, 0)
			session.repeat(ctx, func() error {
				_, err := session.api.Pass(record)
				return err
			})
		}
	}
}

func convert(record *tindergo.RecsCoreUser) types.Person {
	person := types.Person{UserID: record.ID, Name: record.Name, Bio: record.Bio}
	person.Photo = make([]string, 0)
	for _, photo := range record.Photos {
		person.Photo = append(person.Photo, photo.URL)
	}

	return person
}

func (session *tinderSearchSession) appendResult(person *types.Person) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.results = append(session.results, person)
}
