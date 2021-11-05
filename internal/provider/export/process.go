package export

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
)

const (
	batchSize     = 10
	delayMin      = 100 * time.Millisecond
	delayMax      = 150 * time.Millisecond
	delayErrorMin = 10 * time.Second
	delayErrorMax = 5 * time.Minute
)

func (session *exportSession) process(ctx context.Context) {
	session.log.Debugf("Starting Export Session....")

	session.mutex.Lock()
	session.status = types.StatusRunning
	session.mutex.Unlock()

	var file *os.File
	var err error
	if utils.IsFileExists(session.state.FileName) {
		file, err = os.OpenFile(session.state.FileName, os.O_RDWR, os.ModePerm)
		if err == nil {
			file.Seek(-1024, io.SeekEnd)
		}
	} else {
		file, err = os.Create(session.state.FileName)
	}

	if err != nil {
		session.raise(err)
		return
	}

	defer file.Close()

	session.tw = tar.NewWriter(file)
	defer session.tw.Close()

	session.log.Infof("Export to file: %s", session.state.FileName)

	for {
		utils.Delay(ctx, utils.Range{Min: delayMin, Max: delayMax})
		results, total, err := storage.LoadPersons(session.state.Search.taskID, false,
			batchSize, session.state.Offset, session.state.Search.ExportFavourite, 0)
		if err != nil {
			session.raise(err)
			return
		}

		session.total = total

		if len(results) == 0 {
			session.mutex.Lock()
			session.status = types.StatusDone
			session.mutex.Unlock()
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

		session.log.Debugf("Export statistic: %+v", &session.state.Stat)
	}
}

func (session *exportSession) processResult(ctx context.Context, result *types.PersonRecord) {
	if session.state.Search.ExportPhotos {
		for i, url := range result.Person.Photo {
			photo, err := utils.HttpRequest(url)
			if err != nil {
				session.raise(err)
				continue
			}

			fileName := fmt.Sprintf("p%d_%d.jpg", result.ID, i)
			session.repeat(ctx, func() error {
				return session.append(fileName, photo)
			})

			atomic.AddUint32(&session.state.Stat.Processed, uint32(len(photo)))
		}
	}

	if session.state.Search.ExportAbout {
		fileName := fmt.Sprintf("p%d_bio.txt", result.ID)
		session.repeat(ctx, func() error {
			return session.append(fileName, []byte(result.Person.Bio))
		})

		atomic.AddUint32(&session.state.Stat.Processed, uint32(len(result.Person.Bio)))
	}

	if session.state.Search.ExportDump {
		fileName := fmt.Sprintf("p%d.json", result.ID)
		data, err := json.Marshal(&result.Person)
		if err != nil {
			session.log.Warningf("Dump error: %+v", err)
			return
		}
		session.repeat(ctx, func() error {
			return session.append(fileName, data)
		})

		atomic.AddUint32(&session.state.Stat.Processed, uint32(len(data)))
	}
}

func (session *exportSession) append(fileName string, content []byte) error {
	header := new(tar.Header)
	header.Name = fileName
	header.Size = int64(len(content))
	header.Mode = 0644
	if err := session.tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(session.tw, bytes.NewBuffer(content)); err != nil {
		return err
	}

	return nil
}

func (session *exportSession) repeat(ctx context.Context, handler func() error) {
	for {
		err := handler()
		if err == nil {
			return
		}

		session.raise(err)

		utils.Delay(ctx, utils.Range{Min: delayErrorMin, Max: delayErrorMax})
	}
}
