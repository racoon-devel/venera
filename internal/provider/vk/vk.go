package vk

import (
	"github.com/gorilla/mux"
	"github.com/racoon-devel/venera/internal/types"
	"html/template"
	"net/http"
)

import (
	"github.com/ccding/go-logging/logging"
)

type searchSettings struct {
	types.SearchSettings
	Login    string
	Password string
	City     string
}

type Provider struct {
	log *logging.Logger
}

func (provider Provider) ID() string {
	return "vk"
}

func (provider Provider) SetupRouter(router *mux.Router) {
}

func (provider *Provider) SetLogger(log *logging.Logger) {
	provider.log = log
}

func (provider Provider) CreateSearchSession(r *http.Request) (types.SearchSession, error) {
	settings, err := parseForm(r, false)
	if err != nil {
		return nil, err
	}

	if err := settings.SearchSettings.Validate(); err != nil {
		return nil, err
	}

	return &searchSession{
		provider: provider,
		log:      provider.log,
		state:    sessionState{Search: *settings},
	}, nil
}

func (provider Provider) RestoreSearchSession(state string) types.SearchSession {
	session := &searchSession{
		provider: provider,
		log:      provider.log,
	}
	session.LoadState(state)
	return session
}

func (provider Provider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Open",
			Link:  template.URL(result.Person.Link),
		},
	}
}
