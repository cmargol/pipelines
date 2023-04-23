package pipelines

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type fatalErr struct {
	err error
}

func (fe fatalErr) Fatal() string { return "" }
func (fe fatalErr) Error() string { return fe.err.Error() }

func TestWorkerFunctionErrWrapper(t *testing.T) {
	service := "TestService"
	stage := "TestStage"

	// Test with a function that does not return an error
	fNoError := func(s string) (string, error) {
		return s, nil
	}
	wrappedNoError := WorkerFunctionErrWrapper(fNoError, service, stage)
	res, err := wrappedNoError("test")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if res != "test" {
		t.Errorf("expected 'test', got: %v", res)
	}

	// Test with a function that returns a non-ErrFatal error
	fNonFatalError := func(s string) (string, error) {
		return "", errors.New("non-fatal error")
	}
	wrappedNonFatalError := WorkerFunctionErrWrapper(fNonFatalError, service, stage)
	_, err = wrappedNonFatalError("test")
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if _, ok := err.(PipelineErr); !ok {
		t.Errorf("expected a PipelineErr, got: %T", err)
	}

	// Test with a function that returns an ErrFatal error
	fFatalError := func(s string) (string, error) {
		return "", fatalErr{fmt.Errorf("fatal err")}
	}
	wrappedFatalError := WorkerFunctionErrWrapper(fFatalError, service, stage)
	_, err = wrappedFatalError("test")
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if _, ok := err.(ErrFatal); !ok {
		t.Errorf("expected an ErrFatal, got: %T", err)
	}
}

func TestQueueFunctionErrWrapper(t *testing.T) {
	service := "TestService"
	stage := "TestStage"
	ctx := context.Background()

	// Test with a function that does not return an error
	fNoError := func(ctx context.Context) (string, error) {
		return "test", nil
	}
	wrappedNoError := QueueFunctionErrWrapper(fNoError, service, stage)
	res, err := wrappedNoError(ctx)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if res != "test" {
		t.Errorf("expected 'test', got: %v", res)
	}

	// Test with a function that returns a non-ErrFatal error
	fNonFatalError := func(ctx context.Context) (string, error) {
		return "", errors.New("non-fatal error")
	}
	wrappedNonFatalError := QueueFunctionErrWrapper(fNonFatalError, service, stage)
	_, err = wrappedNonFatalError(ctx)
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if _, ok := err.(PipelineErr); !ok {
		t.Errorf("expected a PipelineErr, got: %T", err)
	}

	// Test with a function that returns an ErrFatal error
	fFatalError := func(ctx context.Context) (string, error) {
		return "", fatalErr{err: fmt.Errorf("fatal error")}
	}
	wrappedFatalError := QueueFunctionErrWrapper(fFatalError, service, stage)
	_, err = wrappedFatalError(ctx)
	if err == nil {
		t.Error("expected an error, got nil")
		if err == nil {
			t.Error("expected an error, got nil")
		}
		if _, ok := err.(ErrFatal); !ok {
			t.Errorf("expected an ErrFatal, got: %T", err)
		}
	}
}

func TestDequeueFunctionErrWrapper(t *testing.T) {
	service := "TestService"
	stage := "TestStage"

	// Test with a function that does not return an error
	fNoError := func(s string) error {
		return nil
	}
	wrappedNoError := DequeueFunctionErrWrapper(fNoError, service, stage)
	err := wrappedNoError("test")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Test with a function that returns a non-ErrFatal error
	fNonFatalError := func(s string) error {
		return errors.New("non-fatal error")
	}
	wrappedNonFatalError := DequeueFunctionErrWrapper(fNonFatalError, service, stage)
	err = wrappedNonFatalError("test")
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if _, ok := err.(PipelineErr); !ok {
		t.Errorf("expected a PipelineErr, got: %T", err)
	}

	// Test with a function that returns an ErrFatal error
	fFatalError := func(s string) error {
		return fatalErr{err: fmt.Errorf("fatal error")}
	}
	wrappedFatalError := DequeueFunctionErrWrapper(fFatalError, service, stage)
	err = wrappedFatalError("test")
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if _, ok := err.(ErrFatal); !ok {
		t.Errorf("expected an ErrFatal, got: %T", err)
	}
}
