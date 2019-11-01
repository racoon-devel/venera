package mamba

import (
	"context"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const (
	delayBatchMin = 3 * time.Minute
	delayBatchMax = 5 * time.Minute

	mambaAppID     uint = 2341
	mambaSecretKey      = "3Y3vnn573vt2S4tl6lW8"
)

func (session *mambaSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Mamba API Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.api = newMambaRequester(mambaAppID, mambaSecretKey)
	session.mutex.Unlock()

	// TODO: search
}
