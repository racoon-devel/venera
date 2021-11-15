package vk

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"strconv"
	"sync/atomic"
	"time"
)

func (session *searchSession) process(ctx context.Context) {
	session.log.Debugf("VK starting session...")
	session.ctx = ctx

	if err := session.checkAuth(); err != nil {
		session.raise(err)
		return
	}

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.api = api.NewVK(session.state.AccessToken)
	session.api.Limit = api.LimitUserToken
	session.rater = rater.NewRater(session.state.Search.Rater, "vk", session.log, &session.state.Search.SearchSettings)
	defer session.rater.Close()
	session.mutex.Unlock()

	for session.status != types.StatusError {
		switch session.state.State {
		case stateInitialize:
			session.initialize()
		case stateUserSearch:
			session.userSearch()
		case stateNameUserSearch:
			session.nameUserSearch()
		case stateGroupSearch:
			session.groupSearch()
		case stateSearchInGroups:
			session.searchInGroups()
		default:
			panic("not implemented")
		}

		session.checkStop()
	}
}

func (session *searchSession) initialize() {
	var country, city int
	session.try(func() error {
		var err error
		country, city, err = session.getLocationIDs()
		return err
	})

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.state.CommonData = commonData{
		CountryID: country,
		CityID:    city,
	}
	session.state.State = stateUserSearch
	session.log.Info("[vk] step 1. Global searching...")
}

func (session *searchSession) userSearch() {
	state := &session.state.UserSearch
	p := api.Params{
		"country":   session.state.CommonData.CountryID,
		"city":      session.state.CommonData.CityID,
		"sex":       sexWoman,
		"status":    statusActiveSearch,
		"age_from":  session.state.Search.AgeFrom,
		"age_to":    session.state.Search.AgeTo,
		"has_photo": 1,
		"sort":      state.ReverseSort,
		"count":     searchRequestBatch,
		"offset":    state.Offset,
		"fields":    searchFields,
	}

	session.log.Debugf("[vk] global searching request, offset = %d, reverse sort = %b", state.Offset, state.ReverseSort)

	var resp api.UsersSearchResponse
	session.try(func() error {
		var err error
		resp, err = session.api.UsersSearch(p)
		return err
	})

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user)
	}

	session.mutex.Lock()
	if state.Offset >= vkSearchResultsLimit || len(resp.Items) == 0 {
		if !state.ReverseSort {
			state.ReverseSort = true
			state.Offset = 0
		} else {
			session.state.State = stateNameUserSearch
			session.log.Info("[vk] step 2. Searching by name...")
		}
	}
	session.mutex.Unlock()
}

func (session *searchSession) nameUserSearch() {
	state := &session.state.NameUserSearch
	p := api.Params{
		"q":         womenNames[state.NameIndex],
		"country":   session.state.CommonData.CountryID,
		"city":      session.state.CommonData.CityID,
		"sex":       sexWoman,
		"status":    statusActiveSearch,
		"age_from":  session.state.Search.AgeFrom,
		"age_to":    session.state.Search.AgeTo,
		"has_photo": 1,
		"sort":      state.ReverseSort,
		"count":     searchRequestBatch,
		"offset":    state.Offset,
		"fields":    searchFields,
	}

	session.log.Debugf("[vk] search by name '%s' request, offset = %d, reverse sort = %b", womenNames[state.NameIndex], state.Offset, state.ReverseSort)

	var resp api.UsersSearchResponse
	session.try(func() error {
		var err error
		resp, err = session.api.UsersSearch(p)
		return err
	})

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user)
	}

	session.mutex.Lock()
	if state.Offset >= vkSearchResultsLimit || len(resp.Items) == 0 {
		state.Offset = 0
		if !state.ReverseSort {
			state.ReverseSort = true
		} else {
			state.NameIndex++
			if state.NameIndex >= len(womenNames) {
				session.state.State = stateGroupSearch
				session.log.Info("[vk] step 3. Fetching groups...")
			}
		}
	}
	session.mutex.Unlock()
}

func (session *searchSession) groupSearch() {
	state := &session.state.GroupSearch
	p := api.Params{
		"q":          session.state.Search.Keywords[state.KeywordIndex],
		"type":       "group",
		"country_id": session.state.CommonData.CountryID,
		"offset":     state.Offset,
		"count":      maxGroupSearchLimit,
	}

	session.log.Debugf("[vk] search groups by '%s' request", session.state.Search.Keywords[state.KeywordIndex])
	var resp api.GroupsSearchResponse
	session.try(func() error {
		var err error
		resp, err = session.api.GroupsSearch(p)
		return err
	})

	session.log.Debugf("[vk] retrieved groups: %d", len(resp.Items))

	for _, group := range resp.Items {
		if group.IsClosed == 1 {
			continue
		}
		session.log.Debugf("[vk] found group '%s'", group.Name)
		groupAdd(state, &group)
		atomic.AddUint32(&session.state.Stat.Groups, 1)
	}

	session.mutex.Lock()
	state.Offset += len(resp.Items)
	if state.Offset >= maxGroupSearchLimit || len(resp.Items) == 0 {
		state.Offset = 0
		state.KeywordIndex++
		if state.KeywordIndex >= len(session.state.Search.Keywords) {
			session.state.State = stateSearchInGroups
			session.log.Info("[vk] step 4. Searching in groups...")
		}
	}
	session.mutex.Unlock()
}

func (session *searchSession) searchInGroups() {
	state := &session.state.InGroupSearch
	p := api.Params{
		"country":   session.state.CommonData.CountryID,
		"city":      session.state.CommonData.CityID,
		"sex":       sexWoman,
		"status":    statusActiveSearch,
		"age_from":  session.state.Search.AgeFrom,
		"age_to":    session.state.Search.AgeTo,
		"has_photo": 1,
		"sort":      state.ReverseSort,
		"count":     searchRequestBatch,
		"offset":    state.Offset,
		"fields":    searchFields,
		"group_id":  session.state.GroupSearch.Groups[state.GroupIndex],
	}

	var resp api.UsersSearchResponse
	session.try(func() error {
		var err error
		resp, err = session.api.UsersSearch(p)
		return err
	})

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user)
	}

	session.mutex.Lock()
	if state.Offset >= vkSearchResultsLimit || len(resp.Items) == 0 {
		state.Offset = 0
		if !state.ReverseSort {
			state.ReverseSort = true
		} else {
			state.GroupIndex++
			if state.GroupIndex >= len(session.state.GroupSearch.Groups) {
				session.state.State = stateFriendsSearch
				session.log.Info("[vk] step 5. Search by friends...")
			}
		}
	}
	session.mutex.Unlock()
}

func (session *searchSession) rateUser(u *object.UsersUser) {
	// private-аккануты не годятся
	if u.IsClosed && !u.CanAccessClosed {
		return
	}

	// фильтруем устаревшие аккаунты
	if time.Since(time.Unix(int64(u.LastSeen.Time), 0)) >= expiredAccountThreshold {
		return
	}

	// если такой пользователь уже есть в базе, игнорируем
	if storage.SearchPerson(session.provider.ID(), strconv.Itoa(u.ID)) != nil {
		return
	}

	atomic.AddUint32(&session.state.Stat.Retrieved, 1)

	person := convertPersonRecord(u)

	//дозагрузка фото
	if session.rater.NeedPhotos() {
		session.loadPhotos(person)
	}

	// смотрим рейтинг
	rating := session.rater.Rate(person)
	if rating > 0 {
		if _, err := storage.AppendPerson(person, session.taskID, session.provider.ID()); err != nil {
			session.log.Errorf("[vk] save person failed: %+v", err)
		} else {
			atomic.AddUint32(&session.state.Stat.Saved, 1)
		}
	}

	session.log.Debugf("rate of '%s': %d", person.Name, rating)

	// проверка, не остановили ли задачу
	session.checkStop()
}
