package badoo

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/racoon-devel/venera/internal/provider/badoo/badoogo"
	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
)

const (
	minProfileDelay = 1 * time.Second
	maxProfileDelay = 20 * time.Second

	minBetweenDelay = 1 * time.Minute
	maxBetweenDelay = 3 * time.Minute

	minSessionInterval = 1 * time.Hour
	maxSessionInterval = 2 * time.Hour

	datingDuration  = 30 * time.Minute
	walkingDuration = 20 * time.Minute
)

func (session *badooSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Badoo Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.rater = rater.NewRater(session.state.Search.Rater, "badoo", session.log, &session.state.Search.SearchSettings)
	session.browser = badoogo.NewBadooRequester(ctx, session.log)
	session.mutex.Unlock()

	session.browser.SetDebug(false)
	defer func() { session.browser.Close() }()
	defer session.rater.Close()

	err := session.repeat(ctx, func() error {
		err := session.browser.Login(session.state.Search.Email,
			session.state.Search.Password,
			session.state.Search.Latitude,
			session.state.Search.Longitude)
		session.browser = session.handleError(ctx, session.browser, err)
		return err
	})

	if err != nil {
		session.raise(err)
		return
	}

	err = session.repeat(ctx, func() error {
		var err error
		session.liker, err = session.browser.Spawn()
		return err
	})

	if err != nil {
		session.raise(err)
		return
	}

	defer func() { session.liker.Close() }()

	err = session.repeat(ctx, func() error {
		var err error
		session.walker, err = session.browser.Spawn()
		return err
	})

	if err != nil {
		session.raise(err)
		return
	}

	defer func() { session.walker.Close() }()

	session.alcoExpr = regexp.MustCompile(`Алкоголь:\n([\p{L} ]+)`)
	session.smokeExpr = regexp.MustCompile(`Курение:\n([\p{L} ]+)`)
	session.bodyExpr = regexp.MustCompile(`(\p{L}+) телосложение`)

	session.work(ctx)
}

func (session *badooSearchSession) work(ctx context.Context) {
	for {
		session.processDating(ctx)

		utils.Delay(ctx, utils.Range{Min: minBetweenDelay,
			Max: maxBetweenDelay})

		session.processWalking(ctx)

		utils.Delay(ctx, utils.Range{Min: minSessionInterval,
			Max: maxSessionInterval})
	}
}

func (session *badooSearchSession) processDating(ctx context.Context) {
	session.log.Info("badoo: dating mode started")
	now := time.Now()

	for time.Now().Sub(now) < datingDuration {
		err := session.liker.Fetch(func(user *badoogo.BadooUser) int {
			person := session.convertPersonRecord(user)
			session.log.Debugf("Person fetched: %+v", &person)
			atomic.AddUint32(&session.state.Stat.Retrieved, 1)

			rating := session.rater.Rate(&person)

			if rating >= session.rater.Threshold(types.LikeThreshold) {
				atomic.AddUint32(&session.state.Stat.Liked, 1)
				session.log.Debugf("Like '%s'", person.Name)
				return badoogo.ActionLike
			}

			session.log.Debugf("Dislike '%s'", person.Name)
			atomic.AddUint32(&session.state.Stat.Disliked, 1)
			return badoogo.ActionPass
		})

		session.liker = session.handleError(ctx, session.liker, err)

		utils.Delay(ctx, utils.Range{Min: minProfileDelay, Max: maxProfileDelay})
	}
}

func (session *badooSearchSession) processWalking(ctx context.Context) {
	session.log.Info("badoo: walking around mode started")
	now := time.Now()

	for time.Now().Sub(now) < walkingDuration {
		err := session.walker.WalkAround(func(user *badoogo.BadooUser) int {
			person := session.convertPersonRecord(user)

			session.log.Debugf("Person fetched: %+v", &person)

			if stored := storage.SearchPerson(session.provider.ID(), person.UserID); stored != nil {
				return badoogo.ActionSkip
			}

			atomic.AddUint32(&session.state.Stat.Retrieved, 1)

			rating := session.rater.Rate(&person)

			if rating > 0 {
				if _, err := storage.AppendPerson(&person, session.taskID, session.provider.ID()); err != nil {
					session.log.Errorf("Save person failed: %+v", err)
				}
			}

			utils.Delay(ctx, utils.Range{Min: minProfileDelay, Max: maxProfileDelay})

			if rating >= session.rater.Threshold(types.SuperLikeThreshold) {
				session.log.Debugf("Like '%s'", person.Name)
				atomic.AddUint32(&session.state.Stat.Liked, 1)
				return badoogo.ActionLike
			}

			return badoogo.ActionSkip
		})

		session.walker = session.handleError(ctx, session.walker, err)

		utils.Delay(ctx, utils.Range{Min: minProfileDelay, Max: maxProfileDelay})
	}
}

func (session *badooSearchSession) convertPersonRecord(record *badoogo.BadooUser) types.Person {
	person := types.Person{
		Name:  record.Name,
		Age:   uint(record.Age),
		Job:   record.Job,
		Photo: record.Photos,
	}

	if record.Interests != "" {
		interests := strings.Replace(record.Interests, "\n", ",", -1)
		person.Bio = fmt.Sprintf("Interests: %s\n%s", interests, record.About)
	} else {
		person.Bio = record.About
	}

	alco := session.alcoExpr.FindStringSubmatch(record.Info)
	if len(alco) >= 2 {
		session.log.Debugf("%+v", alco)
		switch alco[1] {
		case "Много пью":
			person.Alco = types.Positive
		case "Не пью":
			fallthrough
		case "Я против алкоголя":
			person.Alco = types.Negative
		default:
			person.Alco = types.Neutral
		}
	}

	smoke := session.smokeExpr.FindStringSubmatch(record.Info)
	if len(smoke) >= 2 {
		session.log.Debugf("%+v", smoke)
		switch smoke[1] {
		case "Не курю":
			person.Smoke = types.Neutral
		case "Категорически против курения":
			person.Smoke = types.Negative
		default:
			person.Smoke = types.Neutral
		}
	}

	body := session.bodyExpr.FindStringSubmatch(record.Info)
	if len(body) >= 2 {
		session.log.Debugf("%+v", body)
		switch body[1] {
		case "стройное":
			person.Body = types.Thin
		case "полное":
			person.Body = types.Fat
		case "атлетическое":
			person.Body = types.Sport
		default:
			person.Body = types.Neutral
		}
	}

	if strings.Contains(record.Info, "пышка") {
		person.Body = types.Fat
	}

	person.UserID = record.ID
	person.Link = record.URL

	return person
}
