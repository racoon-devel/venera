package tinder

import "racoondev.tk/gitea/racoon/venera/internal/types"

type searchSettings struct {
	User     string
	Password string
	AgeFrom  uint
	AgeTo    uint
	Likes    []string
	Dislikes []string
}

// TinderProvider - provider for searching people in Tinder
type TinderProvider struct {
}

func (ctx TinderProvider) RestoreSearchSession(state string) types.SearchSession {
	var session tinderSearchSession
	session.LoadState(state)
	return &session
}
