package pipelines

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockMetricHandler struct {
	lastSuccessExecuted time.Time
	executionTime       time.Duration
	recordCount         int
	errorCount          int
}

func (mmh *mockMetricHandler) RecordLastSuccessfulExecution(service string, stage string) {
	mmh.lastSuccessExecuted = time.Now()
}
func (mmh *mockMetricHandler) RecordExecutionTime(d time.Duration, service string, stage string, status string) {
	mmh.executionTime = d
}
func (mmh *mockMetricHandler) IncrementRecordCount(service string, stage string) {
	mmh.recordCount++
}
func (mmh *mockMetricHandler) IncrementErrorCount(service string, stage string) {
	mmh.errorCount++
}

func TestMetricWrapperDequeue(t *testing.T) {
	type args[T any] struct {
		f       func(T) error
		service string
		stage   string
		mh      MetricsHandler
	}
	type testCase[T any] struct {
		name            string
		args            args[T]
		concreteMH      *mockMetricHandler
		wantErrCount    int
		wantRecordCount int
	}
	var firstMH = &mockMetricHandler{}
	var secondMH = &mockMetricHandler{}
	tests := []testCase[int]{
		{
			name: "Test metrics are called as expected when there is no error",
			args: args[int]{
				f: func(t int) error {
					return nil
				},
				service: "service",
				stage:   "stage",
				mh:      firstMH,
			},
			concreteMH:      firstMH,
			wantErrCount:    0,
			wantRecordCount: 1,
		},
		{
			name: "Test metrics are updated as expected when there is an error",
			args: args[int]{
				f: func(t int) error {
					return fmt.Errorf("an error")
				},
				service: "service",
				stage:   "stage",
				mh:      secondMH,
			},
			concreteMH:      secondMH,
			wantErrCount:    1,
			wantRecordCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MetricWrapperDequeue(tt.args.f, "service", "stage", tt.args.mh)(1)
			if tt.wantErrCount > 0 && err == nil {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.recordCount != tt.wantRecordCount {
				t.Errorf("got record count = %d , want = %d", tt.concreteMH.recordCount, tt.wantRecordCount)
			}
			if tt.concreteMH.errorCount != tt.wantErrCount {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.executionTime <= 0 {
				t.Errorf("want positve execution time, got : %s", tt.concreteMH.executionTime.String())
			}
			if tt.wantErrCount == 0 && tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time want non zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
			if tt.wantErrCount > 0 && !tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time should be zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
		})
	}
}

func TestMetricWrapperQueue(t *testing.T) {
	type args[T any] struct {
		f       func(ctx context.Context) (T, error)
		service string
		stage   string
		mh      MetricsHandler
	}
	type testCase[T any] struct {
		name            string
		args            args[T]
		concreteMH      *mockMetricHandler
		wantErrCount    int
		wantRecordCount int
	}
	var firstMH = &mockMetricHandler{}
	var secondMH = &mockMetricHandler{}
	tests := []testCase[int]{
		{
			name: "Test metrics are called as expected when there is no error",
			args: args[int]{
				f: func(ctx context.Context) (int, error) {
					return 0, nil
				},
				service: "service",
				stage:   "stage",
				mh:      firstMH,
			},
			concreteMH:      firstMH,
			wantErrCount:    0,
			wantRecordCount: 1,
		},
		{
			name: "Test metrics are updated as expected when there is an error",
			args: args[int]{
				f: func(ctx context.Context) (int, error) {
					return 0, fmt.Errorf("an error")
				},
				service: "service",
				stage:   "stage",
				mh:      secondMH,
			},
			concreteMH:      secondMH,
			wantErrCount:    1,
			wantRecordCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MetricWrapperQueue(tt.args.f, "service", "stage", tt.args.mh)(context.Background())
			if tt.wantErrCount > 0 && err == nil {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.recordCount != tt.wantRecordCount {
				t.Errorf("got record count = %d , want = %d", tt.concreteMH.recordCount, tt.wantRecordCount)
			}
			if tt.concreteMH.errorCount != tt.wantErrCount {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.executionTime <= 0 {
				t.Errorf("want positve execution time, got : %s", tt.concreteMH.executionTime.String())
			}
			if tt.wantErrCount == 0 && tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time want non zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
			if tt.wantErrCount > 0 && !tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time should be zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
		})
	}
}

func TestMetricWrapperWorker(t *testing.T) {
	type args[T1 any, T2 any] struct {
		f       func(T1) (T2, error)
		service string
		stage   string
		mh      MetricsHandler
	}
	type testCase[T any] struct {
		name            string
		args            args[int, int]
		concreteMH      *mockMetricHandler
		wantErrCount    int
		wantRecordCount int
	}
	var firstMH = &mockMetricHandler{}
	var secondMH = &mockMetricHandler{}
	tests := []testCase[int]{
		{
			name: "Test metrics are called as expected when there is no error",
			args: args[int, int]{
				f: func(v int) (int, error) {
					return 0, nil
				},
				service: "service",
				stage:   "stage",
				mh:      firstMH,
			},
			concreteMH:      firstMH,
			wantErrCount:    0,
			wantRecordCount: 1,
		},
		{
			name: "Test metrics are updated as expected when there is an error",
			args: args[int, int]{
				f: func(v int) (int, error) {
					return 0, fmt.Errorf("an error")
				},
				service: "service",
				stage:   "stage",
				mh:      secondMH,
			},
			concreteMH:      secondMH,
			wantErrCount:    1,
			wantRecordCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MetricWrapperWorker(tt.args.f, "service", "stage", tt.args.mh)(1)
			if tt.wantErrCount > 0 && err == nil {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.recordCount != tt.wantRecordCount {
				t.Errorf("got record count = %d , want = %d", tt.concreteMH.recordCount, tt.wantRecordCount)
			}
			if tt.concreteMH.errorCount != tt.wantErrCount {
				t.Errorf("got error count = %d , want = %d", tt.concreteMH.errorCount, tt.wantErrCount)
			}
			if tt.concreteMH.executionTime <= 0 {
				t.Errorf("want positve execution time, got : %s", tt.concreteMH.executionTime.String())
			}
			if tt.wantErrCount == 0 && tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time want non zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
			if tt.wantErrCount > 0 && !tt.concreteMH.lastSuccessExecuted.Equal(time.Time{}) {
				t.Errorf("last succesful time should be zero time, got : %s", tt.concreteMH.lastSuccessExecuted)
			}
		})
	}
}
