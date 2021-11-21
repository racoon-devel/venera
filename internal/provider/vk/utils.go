package vk

import (
	"errors"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/racoon-devel/venera/internal/bot"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
	"strconv"
	"sync/atomic"
	"time"
)

func (session *searchSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("[vk] %+v", err)
	session.status = types.StatusError
	session.lastError = err
	atomic.AddUint32(&session.state.Stat.Errors, 1)
}

func injectCaptcha(captchaText, captchaSID string, params ...api.Params) {
	if len(params) > 0 {
		params[0].CaptchaSID(captchaSID)
		params[0].CaptchaKey(captchaText)
	}
}

func (session *searchSession) createApiEngine() {
	session.api = api.NewVK(session.state.AccessToken)
	session.api.Limit = api.LimitUserToken
	defaultHandler := session.api.Handler

	// обрабатываем общие ошибки, которые не затрагивают состояние сессии
	session.api.Handler = func(method string, params ...api.Params) (api.Response, error) {
		for {
			var e *api.Error
			resp, err := defaultHandler(method, params...)
			if err == nil {
				return resp, err
			}
			if errors.As(err, &e) {
				switch e.Code {

				case api.ErrTooMany:
					utils.Delay(session.ctx, utils.Range{
						Min: time.Second,
						Max: 2 * time.Second,
					})

				case api.ErrFlood:
					session.log.Warn("[vk] flood detection alert, waiting...")
					utils.Delay(session.ctx, utils.Range{
						Min: 1 * time.Minute,
						Max: 1 * time.Hour,
					})

				case api.ErrCaptcha:
					session.log.Warnf("[vk] captcha required: %s %s", e.CaptchaSID, e.CaptchaImg)
					text, err := bot.Request(session.ctx, "Captcha required", e.CaptchaImg)
					if err != nil {
						continue
					}
					injectCaptcha(text, e.CaptchaSID, params...)

				case api.ErrRateLimit:
					session.log.Warn("[vk] request rate limit reached")
					utils.Delay(session.ctx, utils.Range{
						Min: 24 * time.Hour,
						Max: 25 * time.Hour,
					})

				default:
					return resp, err
				}
			} else {
				return resp, err
			}
		}
	}
}

type tryResult int

const (
	trySuccess = iota
	tryRepeat
	tryNext
	tryFailed
)

func (r tryResult) isSuccess() bool {
	return r == trySuccess
}

func (r tryResult) isRepeat() bool {
	return r == tryRepeat
}

func (r tryResult) isNext() bool {
	return r == tryNext
}

func (r tryResult) isFailed() bool {
	return r == tryFailed
}

func (session *searchSession) try(err error) tryResult {
	if err == nil {
		return trySuccess
	}
	var e *api.Error
	if errors.As(err, &e) {
		switch e.Code {
		case api.ErrPermission:
			return tryNext
		case api.ErrUserDeleted:
			return tryNext
		case api.ErrAuth:
			session.log.Warnf("[vk] auth required")
			err = session.signIn()
			if err == nil {
				return tryRepeat
			}
		}
	}

	session.raise(err)
	return tryFailed
}

func (session *searchSession) checkAuth() error {
	if session.state.AccessToken == "" || time.Since(session.state.LastAuthTime) >= 24*time.Hour {
		return session.signIn()
	}

	return nil
}

func (session *searchSession) getLocationIDs() (countryID, cityID int, err error) {
	p := api.Params{
		"code":  "RU",
		"count": 10,
	}
	var countries api.DatabaseGetCountriesResponse
	for {
		countries, err = session.api.DatabaseGetCountries(p)
		r := session.try(err)
		if r.isSuccess() {
			break
		}
		if r.isFailed() {
			return
		}
	}

	if countries.Count == 0 {
		err = errors.New("cannot get country ID")
		return
	}
	countryID = countries.Items[0].ID

	p = api.Params{
		"q":          session.state.Search.City,
		"country_id": strconv.Itoa(countryID),
	}
	var cities api.DatabaseGetCitiesResponse
	for {
		cities, err = session.api.DatabaseGetCities(p)
		r := session.try(err)
		if r.isSuccess() {
			break
		}
		if r.isFailed() {
			return
		}
	}

	if cities.Count == 0 {
		err = errors.New("cannot get city ID")
		return
	}

	cityID = cities.Items[0].ID
	return
}

func (session *searchSession) loadPhotos(p *types.Person) {
	params := api.Params{
		"owner_id": p.UserID,
		"count":    maxPhotosLimit,
	}
	resp, err := session.api.PhotosGetAll(params)
	if err != nil {
		session.log.Errorf("[vk] cannot retrieve photos: %+v", err)
		return
	}

	for _, f := range resp.Items {
		p.Photo = append(p.Photo, f.MaxSize().URL)
	}
}

func groupAdd(state *groupSearch, group *object.GroupsGroup) {
	if state.Groups == nil {
		state.Groups = make([]int, 0)
	}
	found := false
	for i := 0; i < len(state.Groups) && !found; i++ {
		found = state.Groups[i] == group.ID
	}
	if found {
		return
	}

	state.Groups = append(state.Groups, group.ID)
}

func (session *searchSession) checkStop() {
	select {
	case <-session.ctx.Done():
		panic("cancelled")
	case <-time.After(time.Millisecond):
	}
}
