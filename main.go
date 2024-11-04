package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	id   int
	jobs <-chan string
	quit chan bool
	wg   *sync.WaitGroup
}

func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for {
			select {
			case job := <-w.jobs:
				fmt.Printf("Worker %d processing: %s\n", w.id, job)
				time.Sleep(1 * time.Second)
			case <-w.quit:
				fmt.Printf("Worker %d exiting\n", w.id)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.quit <- true
}

type WorkerPool struct {
	workers  []*Worker
	jobs     chan string
	wg       sync.WaitGroup
	mu       sync.Mutex
	workerID int
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		workers: make([]*Worker, 0),
		jobs:    make(chan string),
	}
}

func (p *WorkerPool) AddWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.workerID++
	w := &Worker{
		id:   p.workerID,
		jobs: p.jobs,
		quit: make(chan bool),
		wg:   &p.wg,
	}
	p.workers = append(p.workers, w)
	p.wg.Add(1)
	w.Start()
	fmt.Printf("Added worker %d\n", w.id)
}

func (p *WorkerPool) RemoveWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.workers) == 0 {
		fmt.Println("No workers available to remove")
		return
	}
	w := p.workers[len(p.workers)-1]
	w.Stop()
	p.workers = p.workers[:len(p.workers)-1]
	fmt.Printf("Removed worker %d\n", w.id)
}

func (p *WorkerPool) AddJob(job string) {
	p.jobs <- job
}

func (p *WorkerPool) Stop() {
	p.mu.Lock()
	for _, w := range p.workers {
		w.Stop()
	}
	p.mu.Unlock()
	p.wg.Wait()
	close(p.jobs)
	fmt.Println("Pool shutdown complete")
}

func main() {
	pool := NewWorkerPool()
	pool.AddWorker()
	pool.AddWorker()
	pool.AddWorker()

	testJobs := []string{
		"Task 1",
		"Task 2",
		"Task 3",
		"Task 4",
		"Task 5",
	}

	for _, job := range testJobs {
		pool.AddJob(job)
	}

	time.Sleep(2 * time.Second)
	pool.AddWorker()

	time.Sleep(2 * time.Second)
	pool.RemoveWorker()

	time.Sleep(3 * time.Second)
	pool.Stop()
}
