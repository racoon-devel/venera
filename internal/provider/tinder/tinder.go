package tinder

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type searchSettings struct {
	types.SearchSettings
	Tel       string
	APIToken  string
	RefreshToken string
	Latitude  float32
	Longitude float32
}

// TinderProvider - provider for searching people in Tinder
type TinderProvider struct {
	log *logging.Logger
}

func (provider TinderProvider) newSearchSession(search *searchSettings) *tinderSearchSession {
	return &tinderSearchSession{state: tinderSessionState{Search: *search},
		log: provider.log, provider: provider}
}

func (provider TinderProvider) ID() string {
	return "tinder"
}

func (provider TinderProvider) SetupRouter(router *mux.Router) {
	router.HandleFunc("/login", loginHandler).Methods("GET")
}

func (provider *TinderProvider) SetLogger(log *logging.Logger) {
	provider.log = log
}

func (provider TinderProvider) CreateSearchSession(r *http.Request) (types.SearchSession, error) {
	settings, auth, err := parseForm(r, false)
	if err != nil {
		return nil, err
	}

	if err := settings.SearchSettings.Validate(); err != nil {
		return nil, err
	}

	if err := auth.ValidateCode(auth.LoginCode); err != nil {
		return nil, fmt.Errorf("Tinder auth failed: %+v", err)
	}

	if err := auth.Login(); err != nil {
		return nil, fmt.Errorf("Tinder auth failed: %+v", err)
	}

	settings.APIToken = auth.APIToken
	settings.RefreshToken = auth.RefreshToken

	return provider.newSearchSession(settings), nil
}

func (provider TinderProvider) RestoreSearchSession(state string) types.SearchSession {
	session := provider.newSearchSession(&searchSettings{})
	session.LoadState(state)
	return session
}

func (provider TinderProvider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Superlike",
			Link: template.URL(fmt.Sprintf("/task/%d/superlike?id=%s", result.TaskID,
				result.Person.UserID)),
			Command: fmt.Sprintf("/action %d superlike id %s", result.TaskID,
				result.Person.UserID),
		},
	}
}
