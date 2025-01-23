package test

import (
	"github.com/HiChen85/googg"
	"reflect"
	"testing"
)

type args[T googg.SupportedReverseType] struct {
	s []T
}
type testCase[T googg.SupportedReverseType] struct {
	name string
	args args[T]
	want []T
}

func TestReverseInt(t *testing.T) {

	tests := []testCase[int]{
		{
			name: "reverse integers",
			args: args[int]{s: []int{1, 2, 3}},
			want: []int{3, 2, 1},
		},
		{
			name: "reverse empty slice",
			args: args[int]{s: []int{}},
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := googg.Reverse(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	test2 := []testCase[string]{
		{
			name: "reverse strings",
			args: args[string]{s: []string{"a", "b", "c"}},
			want: []string{"c", "b", "a"},
		},
		{
			name: "reverse single string",
			args: args[string]{s: []string{"a"}},
			want: []string{"a"},
		},
	}
	for _, tt := range test2 {
		t.Run(tt.name, func(t *testing.T) {
			if got := googg.Reverse(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
