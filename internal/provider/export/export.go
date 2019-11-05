package export

import (
	"net/http"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type searchSettings struct {
	types.SearchSettings
	taskID          uint
	ExportFavourite bool
	ExportPhotos    bool
	ExportAbout     bool
	ExportDump      bool
}

type ExportProvider struct {
	log *logging.Logger
}

func (provider ExportProvider) newSession(search *searchSettings) *exportSession {
	return &exportSession{state: exportState{Search: *search},
		log: provider.log, provider: provider}
}

func (provider ExportProvider) ID() string {
	return "export"
}

func (provider ExportProvider) SetupRouter(router *mux.Router) {
}

func (provider *ExportProvider) SetLogger(log *logging.Logger) {
	provider.log = log
}

func (provider ExportProvider) CreateSearchSession(r *http.Request) (types.SearchSession, error) {
	settings, err := parseForm(r)
	if err != nil {
		return nil, err
	}

	return provider.newSession(settings), nil
}

func (provider ExportProvider) RestoreSearchSession(state string) types.SearchSession {
	session := provider.newSession(&searchSettings{})
	session.LoadState(state)
	return session
}

func (provider ExportProvider) GetResultActions(result *types.PersonRecord) []types.Action {
	return []types.Action{}
}
