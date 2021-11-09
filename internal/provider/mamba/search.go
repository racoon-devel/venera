package mamba

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/racoon-devel/venera/internal/storage"

	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/types"
)

const (
	mambaAppID     uint = 2341
	mambaSecretKey      = "3Y3vnn573vt2S4tl6lW8"
)

func (session *mambaSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Mamba API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.api = newMambaRequester(mambaAppID, mambaSecretKey)
	session.rater = rater.NewRater(session.state.Search.Rater, "mamba", session.log, &session.state.Search.SearchSettings)
	session.mutex.Unlock()

	defer session.rater.Close()

	session.lookForExp = regexp.MustCompile(`с парнем в возрасте ([\d]+) - ([\d]+) лет`)

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
			session.mutex.Lock()
			session.state.Offset++
			session.mutex.Unlock()

			continue
		}

		if len(users) == 0 {
			session.log.Infof("Mamba session done. Offset = %d, rerun", session.state.Offset)

			session.mutex.Lock()
			session.state.Offset = 0
			session.mutex.Unlock()

			continue
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

	person := convertPersonRecord(user, photos, visitTime[0])
	if stored := storage.SearchPerson(session.provider.ID(), person.UserID); stored != nil {
		session.log.Debugf("Skip exists person '%s'", person.UserID)
		return
	}

	rating := session.rater.Rate(&person)

	// TODO: optimization
	if !session.checkLookFor(user.Familiarity.LookFor) {
		person.Rating = rater.IgnorePerson
	}

	rating = person.Rating

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

func (session *mambaSearchSession) checkLookFor(lookFor string) bool {
	matches := session.lookForExp.FindStringSubmatch(lookFor)
	if len(matches) == 3 {
		ageFrom, _ := strconv.ParseUint(matches[1], 10, 16)
		ageTo, _ := strconv.ParseUint(matches[2], 10, 16)
		return types.MyAge >= ageFrom && types.MyAge <= ageTo
	}
	return true
}

func convertPersonRecord(record *mambaUser, extraPhotos []string, visitTime time.Time) types.Person {
	person := types.Person{
		UserID:    strconv.Itoa(record.Info.Oid),
		Name:      record.Info.Name,
		Bio:       record.About,
		VisitTime: visitTime,
		Link:      strings.Replace(record.Info.Link, "%PARTNER_URL%", "https://mamba.ru", 1),
	}

	person.Age = uint(record.Info.Age)
	person.Photo = make([]string, 1, len(extraPhotos)+1)
	person.Photo[0] = record.Info.Photo

	if len(extraPhotos) != 0 {
		person.Photo = append(person.Photo, extraPhotos...)
	}

	if record.Flags.IsLeader == 1 || record.Flags.IsReal == 1 || record.Flags.IsVIP == 1 {
		person.VIP = true
	}

	if len(record.Interests) > 0 {
		interests := ""
		for _, interest := range record.Interests {
			interests += interest + " "
		}

		person.Bio = fmt.Sprintf("Interests: %s\n%s", interests, person.Bio)
	}

	if record.TypeDetail.Drink != "" {
		switch record.TypeDetail.Drink {
		case "не пью вообще":
			person.Alco = types.Negative
		case "пью в компаниях изредка":
			person.Alco = types.Neutral
		case "люблю выпить":
			person.Alco = types.Positive
		}
	}

	if record.TypeDetail.Smoke != "" {
		if record.TypeDetail.Smoke != "не курю" {
			person.Smoke = types.Positive
		}
	}

	if record.TypeDetail.Constitution != "" {
		switch record.TypeDetail.Constitution {
		case "полное":
			person.Body = types.Fat
		case "обычное":
		case "худощавое":
			person.Body = types.Thin
		default:
			person.Body = types.Sport

		}
	}

	return person
}
