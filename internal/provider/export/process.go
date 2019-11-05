package export

import (
	"context"
	"racoondev.tk/gitea/racoon/venera/internal/storage"
	"racoondev.tk/gitea/racoon/venera/internal/types"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
	"sync/atomic"
)

const batchSize = 10

func (session *exportSession) process(ctx context.Context) {
	session.log.Debugf("Starting Export Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.mutex.Unlock()

	for {
		results, total, err := storage.LoadPersons(session.state.Search.taskID, false,
			batchSize, session.state.Offset, session.state.Search.ExportFavourite, 0)
		if err != nil {
			session.raise(err)
			return
		}

		session.total = total

		if len(results) == 0 {
			// TODO: done
			return
		}

		for i := range results {
			session.processResult(ctx, &results[i])
		}

		session.mutex.Lock()
		session.state.Offset += uint(len(results))
		session.mutex.Unlock()

		percent := (float32(session.state.Offset) / float32(session.total)) * 100.
		atomic.StoreUint32(&session.state.Stat.Progress, uint32(percent))
		atomic.AddUint32(&session.state.Stat.Retrieved, uint32(len(results)))
	}
}

func (session *exportSession) processResult(ctx context.Context, result *types.PersonRecord) {
	if session.state.Search.ExportPhotos {
		for _, url := range result.Person.Photo {
			// cancel ability
			session.log.Debugf("Downloading image '%s'", url)
			photo, err := utils.HttpRequest(url)
			if err != nil {
				session.raise(err)
				continue
			}

			atomic.AddUint32(&session.state.Stat.Processed, uint32(len(photo)))
		}
	}
}

