package export

import (
	"fmt"
	"net/http"

	"github.com/racoon-devel/venera/internal/utils"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"

	"github.com/racoon-devel/venera/internal/types"
	uuid "github.com/satori/go.uuid"
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
	// Generate target filename
	uid := uuid.NewV4()
	fileName := fmt.Sprintf("%s/%s.tar", utils.Configuration.Directories.Downloads, uid.String())
	return &exportSession{state: exportState{Search: *search, FileName: fileName},
		log: provider.log, provider: provider}
}

func (provider ExportProvider) ID() string {
	return "export"
}

func (provider ExportProvider) SetupRouter(router *mux.Router) {
	router.PathPrefix("/download/").Handler(
		http.StripPrefix("/export/download/",
			http.FileServer(http.Dir(utils.Configuration.Directories.Downloads+"/")),
		),
	)
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
