package worker

import (
	"sync"

	"github.com/Yury132/Golang-Task-2/internal/models"
	"github.com/rs/zerolog"

	"github.com/Yury132/Golang-Task-2/internal/worker/pool"
)

type MediaService interface {
	CreateThumbnail(info *models.InfoForThumbnail) error
	GetTaskForProcessing() (*models.InfoForThumbnail, error)
}

type MediaHandler struct {
	log          zerolog.Logger
	mediaService MediaService
	pool         *pool.Pool
}

func (mh *MediaHandler) Start() {
	mh.pool.RunBackground(mh.createThumbnail)
}

func (mh *MediaHandler) Shutdown() {
	mh.pool.Stop()
}

func (mh *MediaHandler) createThumbnail() {
	info, err := mh.mediaService.GetTaskForProcessing()
	if err != nil {
		mh.log.Error().Err(err).Send()
		return
	}

	var wg = new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = mh.mediaService.CreateThumbnail(info); err != nil {
			mh.log.Error().Err(err).Send()
			return
		}
	}()
	wg.Wait()
}

func New(log zerolog.Logger, mediaService MediaService, workersNum int) *MediaHandler {
	return &MediaHandler{
		log:          log,
		mediaService: mediaService,
		pool:         pool.New(log, workersNum),
	}
}
