package pool

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Worker struct {
	log  zerolog.Logger
	quit chan struct{}

	task func()
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	go func() {
		w.log.Info().Msg("Start")
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				w.task()

			case <-w.quit:
				w.log.Info().Msg("Stopped")
				return
			}
		}
	}()
}

func (w *Worker) SetTask(task func()) {
	w.task = task
}

func (w *Worker) Stop() {
	w.quit <- struct{}{}
}

func NewWorker(log zerolog.Logger) *Worker {
	return &Worker{
		log:  log,
		quit: make(chan struct{}, 1),
	}
}
