package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWorkerPool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WorkerPool Suite")
}

func sleepPrint(args ...interface{}) *Result {
	if len(args) != 2 {
		return &Result{
			Err: errors.New("number of arguments mismatch"),
		}
	}

	d, ok := args[1].(time.Duration)
	if !ok {
		return &Result{
			Err: errors.New("expects args[1] to be type time.Duration"),
		}
	}

	value, ok := args[0].(int)
	if !ok {
		return &Result{
			Err: errors.New("expects args[0] to be type int"),
		}
	}

	time.Sleep(d)
	fmt.Println(value)

	return &Result{Value: value}
}

var _ = Describe("WorkerPool", func() {
	var wp *workerPool
	var numWorkers int

	BeforeEach(func() {
		numWorkers = 4
		wp = NewWorkerPool(numWorkers, 10)
	})

	Describe("Start", func() {
		var ctx context.Context
		var wg sync.WaitGroup

		JustBeforeEach(func() {
			close(wp.Tasks())

			wg.Add(1)
			go func() {
				wp.Start(ctx)
				wg.Done()
			}()
		})

		When("done all tasks normally", func() {
			BeforeEach(func() {
				ctx = context.Background()

				wp.Tasks() <- &Task{Func: sleepPrint, Args: []interface{}{1, time.Millisecond}}
			})

			JustBeforeEach(func() {
				wg.Wait()
			})

			It("should receive results", func() {
				Expect(wp.Results()).To(Receive(Equal(&Result{
					Value: 1,
					Err:   nil,
				})))
				Expect(wp.Results()).To(BeClosed())
			})
		})

		Context("with cancel", func() {
			var cancel context.CancelFunc
			var cancelAfter time.Duration

			BeforeEach(func() {
				ctx, cancel = context.WithCancel(context.Background())

				for i := 0; i < numWorkers*2; i++ {
					wp.Tasks() <- &Task{Func: sleepPrint, Args: []interface{}{1, 400 * time.Millisecond}}
				}
			})

			JustBeforeEach(func() {
				time.Sleep(cancelAfter)
				cancel()

				wg.Wait()
			})

			When("cancel before all jobs done", func() {
				BeforeEach(func() {
					cancelAfter = 50 * time.Millisecond
				})

				It("finishes half of jobs and closes the result channel", func() {
					for i := 0; i < numWorkers; i++ {
						Expect(wp.Results()).To(Receive(Equal(&Result{
							Value: 1,
							Err:   nil,
						})))
					}
					Expect(wp.Results()).To(BeClosed())
				})
			})

			When("cancel after all jobs done", func() {
				BeforeEach(func() {
					cancelAfter = 1 * time.Second
				})

				It("finishes all jobs and closes the result channel", func() {
					for i := 0; i < numWorkers*2; i++ {
						Expect(wp.Results()).To(Receive(Equal(&Result{
							Value: 1,
							Err:   nil,
						})))
					}
					Expect(wp.Results()).To(BeClosed())
				})
			})
		})
	})
})
