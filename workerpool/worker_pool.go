package workerpool

import (
	"context"
	"fmt"
	"sync"
)

type Task struct {
	Func func(args ...interface{}) *Result
	Args []interface{}
}

type Result struct {
	Value interface{}
	Err   error
}

type WorkerPool interface {
	Start(ctx context.Context)
	Tasks() chan *Task
	Results() chan *Result
}

type workerPool struct {
	numWorkers int
	tasks      chan *Task
	results    chan *Result
	wg         *sync.WaitGroup
}

var _ WorkerPool = (*workerPool)(nil)

func NewWorkerPool(numWorkers int, bufferSize int) *workerPool {
	return &workerPool{
		numWorkers: numWorkers,
		tasks:      make(chan *Task, bufferSize),
		results:    make(chan *Result, bufferSize),
		wg:         &sync.WaitGroup{},
	}
}

func (wp *workerPool) Start(ctx context.Context) {
	// TODO: implementation
	//
	// Starts numWorkers of goroutines, wait until all jobs are done.
	// Remember to closed the result channel before exit.
	defer close(wp.Results())
	tasks_len := len(wp.tasks)
	wp.wg.Add(tasks_len)
	for i := 0; i < tasks_len; i++ {
		go wp.run(ctx)
	}
	wp.wg.Wait()
}

func (wp *workerPool) Tasks() chan *Task {
	return wp.tasks
}

func (wp *workerPool) Results() chan *Result {
	return wp.results
}

func (wp *workerPool) run(ctx context.Context) {
	// TODO: implementation
	//
	// Keeps fetching task from the task channel, do the task,
	// then makes sure to exit if context is done.
	defer wp.wg.Done()
	for {
		select {
    case <-ctx.Done():
			fmt.Println("[context is done]")
			return
		case task := <-wp.Tasks():
			result := task.Func(task.Args[0], task.Args[1])
			wp.Results() <- result
			return
		}
	}
}
