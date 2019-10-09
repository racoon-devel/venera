package tinder

import (
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func (ctx *tinderSearchSession) raise(err error) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.log.Criticalf("tinder: %+v", err)
	ctx.status = types.StatusError
}

func (ctx *tinderSearchSession) repeat(interval time.Duration, action func() error) {
	// TODO: stop
	for {
		<-time.After(interval)
		if err := action(); err != nil {
			ctx.log.Warningf("Action failed: %+v", err)
		}
	}
}
