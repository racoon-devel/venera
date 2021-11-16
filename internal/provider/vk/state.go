package vk

import "time"

const (
	// начальное состояние
	stateInitialize = iota

	// просто поиск по людям
	stateUserSearch

	// хитрый поиск по именам
	stateNameUserSearch

	// поиск групп по интересам
	stateGroupSearch

	// поиск в группах
	stateSearchInGroups

	// глубокий поиск
	stateFreeSearch
)

type commonData struct {
	CountryID int
	CityID    int
}

type userSearch struct {
	Offset      uint32
	ReverseSort bool
}

type nameUserSearch struct {
	NameIndex   int
	Offset      uint32
	ReverseSort bool
}

type groupSearch struct {
	KeywordIndex int
	Offset       int
	Groups       []int
}

type inGroupSearch struct {
	GroupIndex  int
	Offset      uint32
	ReverseSort bool
}

type freeSearch struct {
	UserID int
}

type vkStat struct {
	Retrieved uint32
	Saved     uint32
	Errors    uint32
	Groups    uint32
}

type sessionState struct {
	Search       searchSettings
	AccessToken  string
	LastAuthTime time.Time

	State int

	Stat           vkStat
	CommonData     commonData
	UserSearch     userSearch
	NameUserSearch nameUserSearch
	GroupSearch    groupSearch
	InGroupSearch  inGroupSearch
	FreeSearch     freeSearch
}
