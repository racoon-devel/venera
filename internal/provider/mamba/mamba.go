package mamba

import (
	"html/template"
	"net/http"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"github.com/racoon-devel/venera/internal/types"
)

type searchSettings struct {
	types.SearchSettings
	City   string
	CityID int
}

type MambaProvider struct {
	log *logging.Logger
}

func (provider MambaProvider) newSearchSession(search *searchSettings) *mambaSearchSession {
	return &mambaSearchSession{state: mambaSessionState{Search: *search},
		log: provider.log, provider: provider}
}

func (provider MambaProvider) ID() string {
	return "mamba"
}

func (provider MambaProvider) SetupRouter(router *mux.Router) {
}

func (provider *MambaProvider) SetLogger(log *logging.Logger) {
	provider.log = log
}

func (provider MambaProvider) CreateSearchSession(r *http.Request) (types.SearchSession, error) {
	settings, err := parseForm(r, false)
	if err != nil {
		return nil, err
	}

	if err := settings.SearchSettings.Validate(); err != nil {
		return nil, err
	}

	m := newMambaRequester(mambaAppID, mambaSecretKey)
	settings.CityID, err = m.GetCityID(settings.City)
	if err != nil {
		return nil, err
	}

	return provider.newSearchSession(settings), nil
}

func (provider MambaProvider) RestoreSearchSession(state string) types.SearchSession {
	session := provider.newSearchSession(&searchSettings{})
	session.LoadState(state)
	return session
}

func (provider MambaProvider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Open",
			Link:  template.URL(result.Person.Link),
		},
	}
}
