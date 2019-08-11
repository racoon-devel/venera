package types

import "net/http"

// SearchSession - search session of some provider
type SearchSession interface {
	SaveState() string
	LoadState(string) error
}

// Provider - object for searching people in some social network
type Provider interface {
	ShowSearchPage(w http.ResponseWriter)
	GetSearchSession(r *http.Request) (SearchSession, error)
	RestoreSearchSession(state string) SearchSession
}

const (
	StatusIdle = iota
	StatusRunning
	StatusPaused
	StatusStopped
	StatusError
)
