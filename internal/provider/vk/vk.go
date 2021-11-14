package vk

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/racoon-devel/venera/internal/types"
	"html/template"
	"net/http"
)

import (
	"github.com/ccding/go-logging/logging"
)

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
	return nil, errors.New("not implemented")
}

func (provider Provider) RestoreSearchSession(state string) types.SearchSession {
	return nil
}

func (provider Provider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{
		{
			Title: "Open",
			Link:  template.URL(result.Person.Link),
		},
	}
}
