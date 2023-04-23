package pipelines

import (
	"testing"
)

func TestConvertSliceToClosedChannel(t *testing.T) {
	type args[T any] struct {
		s []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
	}
	tests := []testCase[any]{
		{
			name: "Same values from slice can be read from channel in same order",
			args: struct{ s []any }{s: []any{1, 2, 3, 4, 5, 6, 7}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int
			for v := range ConvertSliceToClosedChannel(tt.args.s) {
				if v != tt.args.s[count] {
					t.Errorf("ConvertSliceToClosedChannel() , while iterating = %v, want %v", v, tt.args.s[count])
				}
				count++
			}
		})
	}
}
