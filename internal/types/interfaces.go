package types

import (
	"net/http"

	"github.com/ccding/go-logging/logging"
)

type SessionStatus int

const (
	StatusIdle = iota
	StatusRunning
	StatusPaused
	StatusStopped
	StatusError
)

// SearchSession - search session of some provider
type SearchSession interface {
	Start()
	Stop()
	Reset()

	Status() SessionStatus

	SaveState() string
	LoadState(string) error
}

// Provider - object for searching people in some social network
type Provider interface {
	ShowSearchPage(w http.ResponseWriter)
	GetSearchSession(log *logging.Logger, r *http.Request) (SearchSession, error)
	RestoreSearchSession(log *logging.Logger, state string) SearchSession
}
