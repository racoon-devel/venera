package badoo

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/racoon-devel/venera/internal/provider/badoo/badoogo"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
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
	atomic.AddUint32(&session.state.Stat.Errors, 1)
	session.lastError = err
}

func (session *badooSearchSession) handleError(ctx context.Context, browser *badoogo.BadooRequester, err error) *badoogo.BadooRequester {
	if err == nil {
		session.errorCounter = 0
		return browser
	}

	// Если операция была просто отменена, тогда здесь вылетит паника - и при завершении задачи не будут создаваться скрины
	utils.Delay(ctx, utils.Range{Min: 1 * time.Millisecond, Max: 2 * time.Millisecond})

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

	session.log.Errorf("Unexpected browser error '%s': %+v", uid.String(), err)
	session.errorCounter++

	if session.errorCounter >= maxAttempts-2 {
		session.log.Warning("Recreate badoo browser")
		nextBrowser, err := session.browser.Spawn()
		if err != nil {
			session.log.Errorf("Create browser failed: %+v", err)
			return browser
		}
		browser.Close()
		browser = nextBrowser
	}

	return browser
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

		session.raise(err)

		if attempts > maxAttempts {
			return fmt.Errorf("repeating action error: %+v", err)
		}
	}
}
