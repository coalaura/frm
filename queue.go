package main

import (
	"sync"
)

type Job func(string) error

type Task struct {
	File string
	Job  Job
}

type Queue struct {
	stop  chan struct{}
	tasks chan Task
	err   error
	wg    sync.WaitGroup
	mx    sync.Mutex
}

func NewQueue(workers int) *Queue {
	queue := &Queue{
		stop:  make(chan struct{}),
		tasks: make(chan Task, workers),
	}

	for i := 0; i < workers; i++ {
		queue.wg.Add(1)

		go func() {
			defer queue.wg.Done()

			for {
				select {
				case <-queue.stop:
					return
				case task := <-queue.tasks:
					if err := task.Job(task.File); err != nil {
						queue.mx.Lock()

						if queue.err == nil {
							queue.err = err
						}

						queue.mx.Unlock()
					}
				}
			}
		}()
	}

	return queue
}

func (q *Queue) Work(job Job, file string) error {
	q.mx.Lock()
	err := q.err
	q.mx.Unlock()

	if err != nil {
		return err
	}

	q.tasks <- Task{
		File: file,
		Job:  job,
	}

	return nil
}

func (q *Queue) Stop() error {
	select {
	case <-q.stop:
		return nil // The channel was closed
	default:
	}

	close(q.stop)

	q.wg.Wait()

	close(q.tasks)

	return q.err
}
