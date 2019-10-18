package tinder

import (
	"fmt"
	"html/template"

	"github.com/gorilla/mux"

	"github.com/ccding/go-logging/logging"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type searchSettings struct {
	types.SearchSettings
	Tel      string
	APIToken string
}

// TinderProvider - provider for searching people in Tinder
type TinderProvider struct {
}

func (ctx TinderProvider) SetupRouter(router *mux.Router) {
	router.HandleFunc("/login", loginHandler).Methods("GET")
	router.HandleFunc("/superlike/{user}", superlikeHandler).Methods("GET")
}

func (ctx TinderProvider) RestoreSearchSession(log *logging.Logger, state string) types.SearchSession {
	session := tinderSearchSession{log: log}
	session.LoadState(state)
	return &session
}

func (ctx TinderProvider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Superlike",
			Link: template.URL(fmt.Sprintf("/task/%d/superlike?id=%d", result.TaskID,
				result.Person.UserID)),
		},
	}
}
