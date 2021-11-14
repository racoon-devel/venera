package vk

import "github.com/racoon-devel/venera/internal/types"

func (session *searchSession) raise(err error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.log.Criticalf("tinder: %+v", err)
	session.status = types.StatusError
	session.lastError = err
}
