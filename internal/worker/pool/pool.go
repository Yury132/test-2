package pool

import (
	"sync"

	"github.com/rs/zerolog"
)

type Pool struct {
	workers []*Worker

	log        zerolog.Logger
	workersNum int
	wg         sync.WaitGroup
}

func (p *Pool) RunBackground(f func()) {
	for i := 0; i < p.workersNum; i++ {
		worker := NewWorker(p.log)
		p.workers = append(p.workers, worker)

		worker.SetTask(f)
		worker.Start(&p.wg)
	}
}

func (p *Pool) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
	p.wg.Wait()
}

func New(log zerolog.Logger, workersNum int) *Pool {
	return &Pool{
		log:        log,
		workersNum: workersNum,
	}
}
