package atreugo

import (
	"errors"
	"testing"
)

func Test_panicOnError(t *testing.T) {
	type args struct {
		err  error
		want bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Panic",
			args: args{
				err:  errors.New("TestPanic"),
				want: true,
			},
		},
		{
			name: "NotPanic",
			args: args{
				err:  nil,
				want: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()

				if tt.args.want && r == nil {
					t.Errorf("panicOnError(): '%v', want '%v'", false, tt.args.want)
				} else if !tt.args.want && r != nil {
					t.Errorf("panicOnError(): '%v', want '%v'", true, tt.args.want)
				}
			}()

			panicOnError(tt.args.err)
		})
	}
}

func Test_indexOf(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Found",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Atreugo",
			},
			want: 2,
		},
		{
			name: "NotFound",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Yeah",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := indexOf(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("indexOf(): '%v', want: '%v'", got, tt.want)
			}
		})
	}
}

func Test_include(t *testing.T) {
	type args struct {
		vs []string
		t  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Found",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Atreugo",
			},
			want: true,
		},
		{
			name: "NotFound",
			args: args{
				vs: []string{"savsgio", "development", "Atreugo"},
				t:  "Yeah",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := include(tt.args.vs, tt.args.t); got != tt.want {
				t.Errorf("include() = %v, want %v", got, tt.want)
			}
		})
	}
}
