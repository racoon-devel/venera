package tinder

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	minDelay    = 600 * time.Millisecond
	maxDelay    = 9300 * time.Millisecond
	maxAttempts = 5
)

func (session *tinderSearchSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("tinder: %+v", err)
	session.status = types.StatusError
	session.lastError = err
}

func (session *tinderSearchSession) repeat(ctx context.Context, handler func() error) error {
	attempts := 1
	for {
		utils.Delay(ctx, utils.Range{Min: minDelay * time.Duration(attempts),
			Max: maxDelay * time.Duration(attempts)})

		err := handler()
		if err == nil {
			return nil
		}

		attempts++

		atomic.AddUint32(&session.state.Stat.Errors, 1)

		session.raise(err)

		if attempts > maxAttempts {
			return fmt.Errorf("repeating action error: %+v", err)
		}
	}
}
