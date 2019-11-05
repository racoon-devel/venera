package export

import (
	"context"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func (session *exportSession) process(ctx context.Context) {
	session.log.Debugf("Starting Export Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.mutex.Unlock()
}


