package tinder

import (
	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type searchSettings struct {
	User     string
	Password string
	Token    string
	AgeFrom  uint
	AgeTo    uint
	Likes    []string
	Dislikes []string
}

// TinderProvider - provider for searching people in Tinder
type TinderProvider struct {
}

func (ctx TinderProvider) RestoreSearchSession(log *logging.Logger, state string) types.SearchSession {
	session := tinderSearchSession{log: log}
	session.LoadState(state)
	return &session
}
