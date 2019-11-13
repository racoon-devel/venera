package badoo

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/satori/go.uuid"

	"racoondev.tk/gitea/racoon/venera/internal/provider/badoo/badoogo"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

const (
	minDelay    = 100 * time.Millisecond
	maxDelay    = 900 * time.Millisecond
	maxAttempts = 5
)

func (session *badooSearchSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("badoo: %+v", err)
	session.status = types.StatusError
	session.lastError = err
}

func (session *badooSearchSession) unexpected(browser *badoogo.BadooRequester, err error) {
	if err == nil {
		return
	}

	uid := uuid.NewV4()

	browser.TakeScreenshot(
		fmt.Sprintf("%s/error_%s.png",
			utils.Configuration.Directories.Downloads,
			uid.String(),
		),
	)

	browser.TracePage(
		fmt.Sprintf("%s/error_%s.html",
			utils.Configuration.Directories.Downloads,
			uid.String(),
		),
	)

	session.log.Errorf("Unexpected browser error: %+v", err)
}

func (session *badooSearchSession) repeat(ctx context.Context, handler func() error) error {
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
