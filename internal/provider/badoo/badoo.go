package badoo

import (
	"html/template"
	"net/http"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"github.com/racoon-devel/venera/internal/types"
)

type searchSettings struct {
	types.SearchSettings
	Email     string
	Password  string
	Latitude  float32
	Longitude float32
}

type BadooProvider struct {
	log *logging.Logger
}

func (provider BadooProvider) newSearchSession(search *searchSettings) *badooSearchSession {
	return &badooSearchSession{state: badooSessionState{Search: *search},
		log: provider.log, provider: provider}
}

func (provider BadooProvider) ID() string {
	return "badoo"
}

func (provider BadooProvider) SetupRouter(router *mux.Router) {
}

func (provider *BadooProvider) SetLogger(log *logging.Logger) {
	provider.log = log
}

func (provider BadooProvider) CreateSearchSession(r *http.Request) (types.SearchSession, error) {
	settings, err := parseForm(r, false)
	if err != nil {
		return nil, err
	}

	if err := settings.SearchSettings.Validate(); err != nil {
		return nil, err
	}

	return provider.newSearchSession(settings), nil
}

func (provider BadooProvider) RestoreSearchSession(state string) types.SearchSession {
	session := provider.newSearchSession(&searchSettings{})
	session.LoadState(state)
	return session
}

func (provider BadooProvider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Open",
			Link:  template.URL(result.Person.Link),
		},
	}
}
