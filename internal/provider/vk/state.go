package vk

import "time"

const (
	// просто поиск по людям
	stateUserSearch = iota

	// поиск групп по интересам
	stateGroupSearch

	// поиск в группах
	stateSearchInGroups

	// поиск по друзьям
	stateFriendsSearch

	// глубокий поиск
	stateFreeSearch
)

type userSearch struct {
}

type groupSearch struct {
}

type inGroupSearch struct {
}

type friendSearch struct {
}

type freeSearch struct {
}

type sessionState struct {
	Search        searchSettings
	AccessToken   string
	LastAuthTime  time.Time
	State         int
	UserSearch    userSearch
	GroupSearch   groupSearch
	InGroupSearch inGroupSearch
	FriendSearch  friendSearch
	FreeSearch    freeSearch
}
