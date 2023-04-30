package service

//
// В теории курса приведен пример использования паттерна FanOut, в котором
// распределие задач производиться роунд робин алгоритмом по списку каналов,
// причем в случае если очередной канал еще не освободился, то алгоритм
// заткнется, даже если все остальные каналы будут свободны.
// Чтобы этого избежать, сдесь я сделал очередь из свободных воркеров. При
// получении задачи воркер удаляется из очереди, после выполнения задачи
// возвращается в очередь.
// В основной функции диспетчера startNow() запускаются две горутины: одна следит
// за каналом с входящим job'ами и назначает их первому свободному воркеру,
// вторая следит за освободившимися воркерами и возвращает их в очередь
//

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/model"
)

const maxWorkers = 5

type jobFunc func(ctx context.Context, URL *model.ShortURL, mu *sync.RWMutex) error

type Processor struct {
	jobCh   chan *model.ShortURL
	doneCh  chan *worker
	workers []*worker
	wg      sync.WaitGroup
}

type worker struct {
	name string
	ctx  context.Context
	mu   *sync.RWMutex
	job  jobFunc
}

func (w *worker) processJob(URL *model.ShortURL, done chan *worker, errCh chan<- error) {
	go func() {
		log.Printf("%s: short: %s, URL: %s", w.name, URL.Short, URL.URL)
		err := w.job(w.ctx, URL, w.mu)
		if err != nil {
			errCh <- err
		}
		done <- w
	}()
}

func NewProcessor(ctx context.Context, job jobFunc) *Processor {
	res := &Processor{
		jobCh:   make(chan *model.ShortURL),
		doneCh:  make(chan *worker),
		workers: make([]*worker, maxWorkers),
		wg:      sync.WaitGroup{},
	}
	mu := sync.RWMutex{}
	for ik := 0; ik < maxWorkers; ik++ {
		w := &worker{
			name: fmt.Sprintf("worker %d", ik),
			ctx:  ctx,
			job:  job,
			mu:   &mu,
		}
		res.workers[ik] = w
	}
	return res
}

func (p *Processor) startNow(stopCh chan struct{}, errCh chan<- error) {

	go func() {
		stop := false
		for {
			select {
			case w := <-p.doneCh:
				p.workers = append(p.workers, w)
				p.wg.Done()
				if stop && len(p.workers) == maxWorkers {
					return
				}
			case <-stopCh:
				stop = true
			}
		}
	}()

	go func() {
		for {
			if len(p.workers) > 0 {
				select {
				case job := <-p.jobCh:
					w := p.workers[0]
					p.workers = p.workers[1:]
					w.processJob(job, p.doneCh, errCh)
				case <-stopCh:
					return
				}

			}
		}
	}()
}

func (p *Processor) ProceedWith(jobs []*model.ShortURL) []error {
	res := make([]error, 0)
	stopCh := make(chan struct{})
	errCh := make(chan error)

	//  colecting errors
	var errwg sync.WaitGroup
	errwg.Add(1)
	go func() {
		for e := range errCh {
			log.Println(e)
			res = append(res, e)
		}
		errwg.Done()
	}()

	p.startNow(stopCh, errCh)
	p.wg.Add(len(jobs))
	go func() {
		for _, job := range jobs {
			p.jobCh <- job
		}
	}()
	p.wg.Wait()
	close(stopCh)
	close(errCh)
	errwg.Wait()
	return res
}
