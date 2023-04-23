package pipelines

import (
	"context"
	"time"
)

// MetricsHandler is responsible for actually updating the underlying metrics each function is expected to take in
// the service and stage
type MetricsHandler interface {
	RecordLastSuccessfulExecution(service string, stage string)
	RecordExecutionTime(t time.Duration, service string, stage string, status string)
	IncrementRecordCount(service string, stage string)
	IncrementErrorCount(service string, stage string)
}

// MetricWrapperQueue wraps a queue function and calls functions of a MetricsHandler
func MetricWrapperQueue[T any](f func(ctx context.Context) (T, error), service string, stage string, mh MetricsHandler) func(ctx context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		mh.IncrementRecordCount(service, stage)
		now := time.Now()
		res, err := f(ctx)
		if err != nil {
			mh.IncrementErrorCount(service, stage)
			mh.RecordExecutionTime(time.Since(now), service, stage, "fail")
			return res, err
		}
		mh.RecordLastSuccessfulExecution(service, stage)
		mh.RecordExecutionTime(time.Since(now), service, stage, "success")
		return res, err
	}
}

// MetricWrapperWorker wraps a worker function and calls functions of a MetricsHandler
func MetricWrapperWorker[T1, T2 any](f func(T1) (T2, error), service string, stage string, mh MetricsHandler) func(T1) (T2, error) {
	return func(v T1) (T2, error) {
		mh.IncrementRecordCount(service, stage)
		now := time.Now()
		res, err := f(v)
		if err != nil {
			mh.IncrementErrorCount(service, stage)
			mh.RecordExecutionTime(time.Since(now), service, stage, "fail")
			return res, err
		}
		mh.RecordLastSuccessfulExecution(service, stage)
		mh.RecordExecutionTime(time.Since(now), service, stage, "success")
		return res, nil
	}
}

// MetricWrapperDequeue wraps a dequeue function and calls functions of a MetricsHandler
func MetricWrapperDequeue[T any](f func(T) error, service string, stage string, mh MetricsHandler) func(T) error {
	return func(v T) error {
		mh.IncrementRecordCount(service, stage)
		now := time.Now()
		err := f(v)
		if err != nil {
			mh.IncrementErrorCount(service, stage)
			mh.RecordExecutionTime(time.Since(now), service, stage, "fail")
			return err
		}
		mh.RecordLastSuccessfulExecution(service, stage)
		mh.RecordExecutionTime(time.Since(now), service, stage, "success")
		return nil
	}
}
