package pipelines

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
)

// TestQueue test that was generated by AI
func TestQueue(t *testing.T) {
	ctx := context.Background()
	bufferSize := 1
	workers := 1

	// Test with a queueFunc that does not return an error and provides content
	counter := 0
	fNoError := func(ctx context.Context) (int, error) {
		counter++
		if counter <= 10 {
			return counter, nil
		}
		return 0, ErrQueueEmpty
	}
	queueChan, errorChan := Queue(ctx, fNoError, bufferSize, workers)
	result := make([]int, 0)
	for v := range queueChan {
		result = append(result, v)
	}
	for range errorChan {
		t.Error("expected no errors")
	}
	sort.Ints(result)
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got: %v", expected, result)
	}

	// Test with a queueFunc that returns an ErrQueueEmpty error
	fErrQueueEmpty := func(ctx context.Context) (int, error) {
		return 0, ErrQueueEmpty
	}
	queueChan, errorChan = Queue(ctx, fErrQueueEmpty, bufferSize, workers)
	if _, ok := <-queueChan; ok {
		t.Error("expected queueChan to be closed")
	}
	if _, ok := <-errorChan; ok {
		t.Error("expected errorChan to be closed")
	}

	// Test with a queueFunc that returns a non-ErrQueueEmpty error
	fNonErrQueueEmpty := func(ctx context.Context) (int, error) {
		return 0, errors.New("non-ErrQueueEmpty error")
	}
	_, errorChan = Queue(ctx, fNonErrQueueEmpty, bufferSize, workers)
	errorCount := 0
	for range errorChan {
		errorCount++
		if errorCount == 10 {
			break
		}
	}
	if errorCount != 10 {
		t.Errorf("expected %d errors, got: %d", workers, errorCount)
	}
}

func TestMerge(t *testing.T) {
	// Test with multiple channels that provide content
	t.Logf("test with multiple channels that provide content")
	ch1 := make(chan int, 5)
	ch2 := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		ch1 <- i
		ch2 <- i * 10
	}
	close(ch1)
	close(ch2)

	merged := Merge(ch1, ch2)
	result := make([]int, 0)
	for v := range merged {
		result = append(result, v)
	}
	sort.Ints(result)
	expected := []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got: %v", expected, result)
	}

	// Test with empty channels
	t.Logf("test with empty channels")
	chEmpty1 := make(chan int)
	chEmpty2 := make(chan int)
	close(chEmpty1)
	close(chEmpty2)

	mergedEmpty := Merge(chEmpty1, chEmpty2)
	if _, ok := <-mergedEmpty; ok {
		t.Error("expected mergedEmpty to be closed")
	}
}

func TestBroadcast(t *testing.T) {
	// Test with multiple subscriber channels that receive content
	source := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		source <- i
	}
	close(source)

	subscriber1 := make(chan int, 5)
	subscriber2 := make(chan int, 5)
	defer close(subscriber1)
	defer close(subscriber2)
	Broadcast(source, subscriber1, subscriber2)

	result1 := make([]int, 0)
	for v := range subscriber1 {
		result1 = append(result1, v)
		if len(result1) == 5 {
			break
		}
	}
	result2 := make([]int, 0)
	for v := range subscriber2 {
		result2 = append(result2, v)
		if len(result2) == 5 {
			break
		}
	}

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result1, expected) {
		t.Errorf("expected %v, got: %v", expected, result1)
	}
	if !reflect.DeepEqual(result2, expected) {
		t.Errorf("expected %v, got: %v", expected, result2)
	}

}

func TestWorkerPool(t *testing.T) {
	bufferSize := 5
	workers := 3

	// Test with a workFunc that does not return an error and processes content
	work := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		work <- i
	}
	close(work)

	workFunc := func(n int) (int, error) {
		return n * 2, nil
	}
	resultChan, errorChan := WorkerPool(work, workFunc, bufferSize, workers)
	result := make([]int, 0)
	for v := range resultChan {
		result = append(result, v)
	}
	for range errorChan {
		t.Error("expected no errors")
	}
	sort.Ints(result)
	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got: %v", expected, result)
	}

	// Test with a workFunc that returns an error
	work = make(chan int, 5)
	for i := 1; i <= 5; i++ {
		work <- i
	}
	close(work)

	workFuncErr := func(n int) (int, error) {
		return 0, fmt.Errorf("error occurred")
	}
	resultChan, errorChan = WorkerPool(work, workFuncErr, bufferSize, workers)
	if _, ok := <-resultChan; ok {
		t.Error("expected resultChan to be closed")
	}
	errorCount := 0
	for range errorChan {
		errorCount++
	}
	if errorCount != 5 {
		t.Errorf("expected 5 errors, got: %d", errorCount)
	}

	// Test with an empty input queue
	emptyWork := make(chan int)
	close(emptyWork)

	resultChan, errorChan = WorkerPool(emptyWork, workFunc, bufferSize, workers)
	if _, ok := <-resultChan; ok {
		t.Error("expected resultChan to be closed")
	}
	if _, ok := <-errorChan; ok {
		t.Error("expected errorChan to be closed")
	}
}

func TestWorkerPoolWithZeroValueFilter(t *testing.T) {
	bufferSize := 5
	workers := 3

	// Test with a workFunc that does not return an error, processes content, and returns zero values
	work := make(chan int, 6)
	for i := 0; i <= 5; i++ {
		work <- i
	}
	close(work)

	workFunc := func(n int) (int, error) {
		return n * 2, nil
	}
	resultChan, errorChan := WorkerPoolWithZeroValueFilter(work, workFunc, bufferSize, workers)
	result := make([]int, 0)
	for v := range resultChan {
		result = append(result, v)
	}
	for range errorChan {
		t.Error("expected no errors")
	}
	sort.Ints(result)
	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got: %v", expected, result)
	}

	// Test with a workFunc that returns an error
	work = make(chan int, 5)
	for i := 1; i <= 5; i++ {
		work <- i
	}
	close(work)

	workFuncErr := func(n int) (int, error) {
		return 0, fmt.Errorf("error occurred")
	}
	resultChan, errorChan = WorkerPoolWithZeroValueFilter(work, workFuncErr, bufferSize, workers)
	if _, ok := <-resultChan; ok {
		t.Error("expected resultChan to be closed")
	}
	errorCount := 0
	for range errorChan {
		errorCount++
	}
	if errorCount != 5 {
		t.Errorf("expected 5 errors, got: %d", errorCount)
	}

	// Test with an empty input queue
	emptyWork := make(chan int)
	close(emptyWork)

	resultChan, errorChan = WorkerPoolWithZeroValueFilter(emptyWork, workFunc, bufferSize, workers)
	if _, ok := <-resultChan; ok {
		t.Error("expected resultChan to be closed")
	}
	if _, ok := <-errorChan; ok {
		t.Error("expected errorChan to be closed")
	}
}

func TestDequeue(t *testing.T) {
	bufferSize := 5
	workers := 3

	// Test with a dequeueFunc that does not return an error and processes the input values
	work := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		work <- i
	}
	close(work)

	mu := sync.Mutex{}
	processedValues := make([]int, 0)
	dequeueFunc := func(n int) error {
		// Mutex lock required due to race conditions
		mu.Lock()
		defer mu.Unlock()
		processedValues = append(processedValues, n)
		return nil
	}
	errorChan := Dequeue(work, dequeueFunc, bufferSize, workers)

	for range errorChan {
		t.Error("expected no errors")
	}
	sort.Ints(processedValues)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(processedValues, expected) {
		t.Errorf("expected %v, got: %v", expected, processedValues)
	}

	// Test with a dequeueFunc that returns an error
	work = make(chan int, 5)
	for i := 1; i <= 5; i++ {
		work <- i
	}
	close(work)

	dequeueFuncErr := func(n int) error {
		return fmt.Errorf("error occurred")
	}
	errorChan = Dequeue(work, dequeueFuncErr, bufferSize, workers)

	errorCount := 0
	for range errorChan {
		errorCount++
	}
	if errorCount != 5 {
		t.Errorf("expected 5 errors, got: %d", errorCount)
	}

	// Test with an empty input queue
	emptyWork := make(chan int)
	close(emptyWork)

	errorChan = Dequeue(emptyWork, dequeueFunc, bufferSize, workers)
	if _, ok := <-errorChan; ok {
		t.Error("expected errorChan to be closed")
	}
}
