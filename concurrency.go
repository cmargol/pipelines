package pipelines

import (
	"context"
	"errors"
	"sync"
)

// Queue calls the queueFunc and adds the results to channels it returns. It is expected that the queueFunc produces new
// content with each call. In the case that queueFunc returns an ErrQueueEmpty it will cause the routine to close. The
// channel buffer size will be equal to the bufferSize that is passed in and the goroutines will be equal to the amount
// of workers passed in. It is suggested that the bufferSize to be >= to the amount of goroutines
//
// In order to orchestrate the closing of the passed in channels sync.WaitGroup is used to wait for the worker
// goroutines to be finished.
func Queue[T any](ctx context.Context, queueFunc func(context.Context) (T, error), bufferSize int, workers int) (<-chan T, <-chan error) {
	// Sanity check for bufSize if it is too low we will set it as an unbuffered channel
	if bufferSize < 0 {
		bufferSize = 0
	}

	// Sanity check for workers we require at least one
	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	queueChan := make(chan T, bufferSize)
	errorChan := make(chan error, bufferSize)

	wg.Add(workers)

	// Start as many workers as requested
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done() // Remove one from wg as this exits
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Default: call the queueFunc for the next results and publish them to the queue or err channel
					res, err := queueFunc(ctx)
					if err != nil {
						if errors.Is(err, ErrQueueEmpty) {
							return
						}
						errorChan <- err
						continue
					}
					queueChan <- res
				}
			}
		}()
	}

	// Create a go routine that is waiting for all workers, so it can shut down it's channels
	go func() {
		wg.Wait()
		close(queueChan)
		close(errorChan)
	}()

	return queueChan, errorChan
}

// Merge function converts a list of channels to a single channel by starting a goroutine for each inbound channel
// that copies the values to the sole outbound channel. Once all the output goroutines have been started, merge starts
// one more goroutine to close the outbound channel after all sends on that channel are done.
//
// Sends on a closed channel panic, so it's important to ensure all sends are done before calling close. The
// sync.WaitGroup type provides a simple way to arrange this synchronization:
func Merge[T any](cs ...<-chan T) chan T {
	var wg sync.WaitGroup
	out := make(chan T, len(cs))

	// Start an output goroutine for each input channel in cs. Output copies values from c to out until c is done
	// or closed, then calls wg.Done
	output := func(c <-chan T) {
		defer wg.Done()
		for v := range c {
			select {
			case out <- v:
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are done. This must start after wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Broadcast takes results from one channel and sends it on multiple channels, note that if the subscriber channels meet
// capacity this will be a blocking call
func Broadcast[T any](cs <-chan T, subscribers ...chan<- T) {
	for v := range cs {
		for _, s := range subscribers {
			s <- v
		}
	}
}

// WorkerPool takes in a channel of work and runs a work function over it and sends the results to channels. The resulting
// channels will be buffered based on the passed in buffer size and the amount of routines that are used for this
// will be equal to the amount of workers passed in.
func WorkerPool[T1, T2 any](queue <-chan T1, workFunc func(T1) (T2, error), bufferSize int, workers int) (<-chan T2, <-chan error) {
	// Sanity check to make sure buffer size and workers are at minimum values
	if bufferSize < 0 {
		bufferSize = 0
	}

	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	out := make(chan T2, bufferSize)
	errc := make(chan error, bufferSize)

	wg.Add(workers)
	// Create workers that will call the workFunc
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for work := range queue {
				res, err := workFunc(work)
				if err != nil {
					errc <- err
					continue
				}
				out <- res
			}
		}()
	}

	// Spin up another goroutine to wait until workers are done until closing the channels
	go func() {
		wg.Wait()
		close(out)
		close(errc)
	}()

	return out, errc
}

// WorkerPoolWithZeroValueFilter takes in a channel of work and runs a work function over it and sends the results to
// channels. The resulting channels will be buffered based on the passed in buffer size and the amount of routines that
// are used for this will be equal to the amount of workers passed in. The main differentiation from WorkerPool is that
// any zero values that are returned from the workFunc are not passed forward as a result but instead dropped.
func WorkerPoolWithZeroValueFilter[T1 any, T2 comparable](queue <-chan T1, workFunc func(T1) (T2, error), bufferSize int, workers int) (<-chan T2, <-chan error) {
	// Sanity check to make sure buffer size and workers are at minimum values
	if bufferSize < 0 {
		bufferSize = 0
	}

	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	out := make(chan T2, bufferSize)
	errc := make(chan error, bufferSize)

	wg.Add(workers)
	// Create workers that will call the workFunc
	for i := 0; i < workers; i++ {
		go func() {
			var zeroValOfT2 T2
			defer wg.Done()
			for work := range queue {
				res, err := workFunc(work)
				if err != nil {
					errc <- err
					continue
				}
				if res != zeroValOfT2 {
					out <- res
				}
			}
		}()
	}

	// Spin up another goroutine to wait until workers are done until closing the channels
	go func() {
		wg.Wait()
		close(out)
		close(errc)
	}()

	return out, errc
}

// Dequeue is/are termination worker(s) that end the pipeline. Examples of this may be printing results, storing
// data to an external source, etc.
func Dequeue[T any](queue <-chan T, dequeueFunc func(T) error, bufferSize int, workers int) <-chan error {
	// Sanity check for workers
	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	errChan := make(chan error, bufferSize)
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		// Spin up the workers
		go func() {
			defer wg.Done()
			for val := range queue {
				if err := dequeueFunc(val); err != nil {
					errChan <- err
				}
			}
		}()
	}

	// Wait till all workers are done before we close the errChan
	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}
