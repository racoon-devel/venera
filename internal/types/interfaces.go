package types

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ccding/go-logging/logging"
	"github.com/gorilla/mux"
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
	Process(ctx context.Context)
	Reset()

	Status() SessionStatus
	GetStat() map[string]uint32

	SaveState() string
	LoadState(string) error

	Results() []*Person
	Action(action string, params url.Values) error

	Update(w http.ResponseWriter, r *http.Request) (bool, error)
}

// Provider - object for searching people in some social network
type Provider interface {
	ID() string
	GetSearchSession(log *logging.Logger, r *http.Request) (SearchSession, error)
	RestoreSearchSession(log *logging.Logger, state string) SearchSession
	GetResultActions(result *PersonRecord) []Action
	SetupRouter(router *mux.Router)
}

type Rater interface {
	Init(settings *SearchSettings)
	Rate(person *Person)
}
