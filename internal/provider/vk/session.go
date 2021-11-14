package vk

import (
	"context"
	"encoding/json"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/types"
	"net/http"
	"net/url"
	"sync"
)

type searchSession struct {
	// Защищено мьютексом
	status    types.SessionStatus
	lastError error
	state     sessionState
	mutex     sync.Mutex

	provider Provider
	taskID   uint
	log      *logging.Logger
	rater    types.Rater

	api *api.VK
}

func (session *searchSession) Process(ctx context.Context, taskID uint) {
	defer func() {
		if r := recover(); r != nil {
			session.log.Errorf("Vk session panic: %+v. Recovered", r)

			session.mutex.Lock()
			defer session.mutex.Unlock()
			session.status = types.StatusStopped
		}
	}()

	session.taskID = taskID
	session.process(ctx)
}

func (session *searchSession) Reset() {
	panic("implement me")
}

func (session *searchSession) Status() types.SessionStatus {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.status
}

func (session *searchSession) GetStat() map[string]uint32 {
	panic("implement me")
}

func (session *searchSession) GetLastError() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	err := session.lastError
	session.lastError = nil
	return err
}

func (session *searchSession) SaveState() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	data, err := json.Marshal(&session.state)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (session *searchSession) LoadState(state string) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	err := json.Unmarshal([]byte(state), &session.state)
	return err
}

func (session *searchSession) Poll() {

}

func (session *searchSession) Action(action string, params url.Values) error {
	panic("implement me")
}

func (session *searchSession) Update(w http.ResponseWriter, r *http.Request) (bool, error) {
	panic("implement me")
}
