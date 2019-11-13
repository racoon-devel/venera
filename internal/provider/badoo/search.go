package badoo

import (
	"context"

	"racoondev.tk/gitea/racoon/venera/internal/rater"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

func (session *badooSearchSession) process(ctx context.Context) {
	session.log.Debugf("Starting Badoo Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.rater = rater.NewRater(session.state.Search.Rater, "badoo", session.log, &session.state.Search.SearchSettings)
	session.mutex.Unlock()

	defer session.rater.Close()
}
