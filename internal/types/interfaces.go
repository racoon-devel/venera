package types

import (
	"context"
	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

type SessionStatus int

const (
	StatusIdle = iota
	StatusRunning
	StatusStopped
	StatusError
)

// SearchSession - search session of some provider
type SearchSession interface {
	Process(ctx context.Context, taskID uint)
	Reset()

	Status() SessionStatus
	GetStat() map[string]uint32
	GetLastError() error

	SaveState() string
	LoadState(string) error

	Poll()
	Action(action string, params url.Values) error
	Update(w http.ResponseWriter, r *http.Request) (bool, error)
}

type Provider interface {
	ID() string

	SetLogger(log *logging.Logger)
	SetupRouter(router *mux.Router)

	CreateSearchSession(r *http.Request) (SearchSession, error)
	RestoreSearchSession(state string) SearchSession

	GetResultActions(result *PersonRecord) []Action
}

type Rater interface {
	Init(log *logging.Logger, settings *SearchSettings)
	Rate(person *Person) int
}
