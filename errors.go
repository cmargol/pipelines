package pipelines

import (
	"context"
	"fmt"
)

type ErrFatal interface {
	Fatal() string
	error
}

type ErrPipeline interface {
	Service() string
	Stage() string
	error
}

var ErrQueueEmpty = fmt.Errorf("no additional values can be added to the queue")

type PipelineErr struct {
	err     error
	service string
	stage   string
}

func NewPipelineErr(err error, service string, stage string) PipelineErr {
	return PipelineErr{
		err:     err,
		service: service,
		stage:   stage,
	}
}

func (e PipelineErr) Error() string {
	return e.Error()
}

func (e PipelineErr) Service() string {
	return e.service
}

func (e PipelineErr) Stage() string {
	return e.stage
}

// WorkerFunctionErrWrapper will wrap a given pipeline function and return the same function but will change the
// error into a PipelineErr if it is not a ErrFatal error
func WorkerFunctionErrWrapper[T1, T2 any](f func(T1) (T2, error), service string, stage string) func(T1) (T2, error) {
	return func(v T1) (T2, error) {
		res, err := f(v)
		if err != nil {
			switch e := err.(type) {
			case ErrFatal:
				return res, e
			}
			return res, NewPipelineErr(err, service, stage)
		}
		return res, nil
	}
}

// QueueFunctionErrWrapper will wrap a given pipeline queue function and return the same function but will change the
// error into a PipelineErr if it is not a ErrFatal error
func QueueFunctionErrWrapper[T any](f func(ctx context.Context) (T, error), service string, stage string) func(ctx context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		res, err := f(ctx)
		if err != nil {
			switch e := err.(type) {
			case ErrFatal:
				return res, e
			}
			return res, NewPipelineErr(err, service, stage)
		}
		return res, nil
	}
}

// DequeueFunctionErrWrapper will wrap a give pipeline queue function and return the same function but will change the
// error into a PipelineErr if it is not a ErrFatal error
func DequeueFunctionErrWrapper[T any](f func(T) error, service string, stage string) func(T) error {
	return func(v T) error {
		if err := f(v); err != nil {
			switch e := err.(type) {
			case ErrFatal:
				return e
			}
			return NewPipelineErr(err, service, stage)
		}
		return nil
	}
}
