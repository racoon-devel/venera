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
	session.createApiEngine()
	session.rater = rater.NewRater(session.state.Search.Rater, "vk", session.log, &session.state.Search.SearchSettings)
	defer session.rater.Close()
	session.mutex.Unlock()

	for session.status == types.StatusRunning {
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
		case stateFreeSearch:
			session.freeSearch()
		default:
			panic("not implemented")
		}

		session.checkStop()
	}
}

func (session *searchSession) initialize() {
	country, city, err := session.getLocationIDs()

	if err != nil {
		return
	}

	if err := storage.RawDB().AutoMigrate(&vkUserRecord{}).Error; err != nil {
		session.raise(err)
		return
	}

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

	resp, err := session.api.UsersSearch(p)
	if err != nil {
		return
	}

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user, true)
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

	resp, err := session.api.UsersSearch(p)
	if err != nil {
		return
	}

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user, true)
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

	resp, err := session.api.GroupsSearch(p)
	if err != nil {
		return
	}

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

	resp, err := session.api.UsersSearch(p)
	if err != nil {
		return
	}

	session.log.Debugf("[vk] found users: %d", len(resp.Items))

	for _, user := range resp.Items {
		atomic.AddUint32(&state.Offset, 1)
		session.rateUser(&user, true)
	}

	session.mutex.Lock()
	if state.Offset >= vkSearchResultsLimit || len(resp.Items) == 0 {
		state.Offset = 0
		if !state.ReverseSort {
			state.ReverseSort = true
		} else {
			state.GroupIndex++
			if state.GroupIndex >= len(session.state.GroupSearch.Groups) {
				session.state.State = stateFreeSearch
				session.log.Info("[vk] step 5. Free search...")
			}
		}
	}
	session.mutex.Unlock()
}

func (session *searchSession) freeSearch() {
	state := &session.state.FreeSearch
	userID := state.UserID
	p := api.Params{
		"fields": "sex,bdate,city,last_seen,relation",
	}
	if state.UserID != 0 {
		p["user_id"] = userID
	}

	resp, err := session.api.FriendsGetFields(p)
	if err != nil {
		return
	}

	session.log.Debugf("[vk] friends fetched: %d", len(resp.Items))

	for _, friend := range resp.Items {
		if !friend.IsClosed && friend.City.ID == session.state.CommonData.CityID {
			if session.isRateableUser(&friend.UsersUser) && !dbContains(friend.ID) {
				session.fetchUser(friend.ID)
			}
			dbAdd(friend.ID)
		}
	}

	if userID != 0 {
		dbRemove(state.UserID)
	}

	userID, ok := dbFetchFirst()
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if !ok {
		session.status = types.StatusDone
		return
	}
	state.UserID = userID
}

func (session *searchSession) fetchUser(ID int) {
	p := api.Params{
		"user_ids": ID,
		"fields":   searchFields + ",counters",
	}

	resp, err := session.api.UsersGet(p)
	if err != nil {
		return
	}

	if resp[0].Counters.Friends >= friendsLimitThreshold {
		session.log.Warnf("[vk] skip person '%s %s', because friends limit reached [ %d > %d ]", resp[0].FirstName, resp[0].LastName, resp[0].Counters.Friends, friendsLimitThreshold)
		return
	}

	session.rateUser(&resp[0], false)
}

func (session *searchSession) rateUser(u *object.UsersUser, checkFriends bool) {
	userName := u.FirstName + " " + u.LastName
	// private-аккануты не годятся
	if u.IsClosed && !u.CanAccessClosed {
		session.log.Debugf("[vk] skip person '%s', because account has private profile", userName)
		return
	}

	// фильтруем устаревшие аккаунты
	if time.Since(time.Unix(int64(u.LastSeen.Time), 0)) >= expiredAccountThreshold {
		session.log.Debugf("[vk] skip person '%s', because account is expired", userName)
		return
	}

	// если такой пользователь уже есть в базе, игнорируем
	if storage.SearchPerson(session.provider.ID(), strconv.Itoa(u.ID)) != nil {
		session.log.Debugf("[vk] skip person '%s', because already rated", userName)
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

		// проверим, а не блогерша ли
		if checkFriends {
			p := api.Params{
				"user_ids": u.ID,
				"fields":   "counters",
			}

			resp, err := session.api.UsersGet(p)
			if err == nil && len(resp) > 0 {
				if resp[0].Counters.Friends > friendsLimitThreshold {
					session.log.Warnf("[vk] skip person '%s %s', because friends limit reached [ %d > %d ]", resp[0].FirstName, resp[0].LastName, resp[0].Counters.Friends, friendsLimitThreshold)
					return
				}
			}
		}
		if _, err := storage.AppendPerson(person, session.taskID, session.provider.ID()); err != nil {
			session.log.Errorf("[vk] save person failed: %+v", err)
		} else {
			atomic.AddUint32(&session.state.Stat.Saved, 1)
		}
	}

	session.log.Debugf("[vk] '%s' rating: %d", person.Name, rating)

	// проверка, не остановили ли задачу
	session.checkStop()
}
