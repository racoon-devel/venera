package tinder

import (
	"context"
	"sync/atomic"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	minDelayMs uint32 = 600
	maxDelayMs uint32 = 5300
)

func (session *tinderSearchSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("tinder: %+v", err)
	session.status = types.StatusError
}

func (session *tinderSearchSession) repeat(ctx context.Context, handler func() error) {
	for {
		utils.Delay(ctx, utils.Range{MinMs: minDelayMs, MaxMs: maxDelayMs})

		err := handler()
		if err == nil {
			return
		}

		atomic.AddUint32(&session.state.Stat.Errors, 1)

		session.raise(err)
	}
}
