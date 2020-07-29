package valuediff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestDeepDiff(t *testing.T) {
	type args struct {
		x interface{}
		y interface{}
	}
	tests := []struct {
		name string
		args args
		want []Diff
	}{
		{
			args: args{
				djson(`1`),
				djson(`1`),
			},
		},
		{
			args: args{
				djson(`0`),
				djson(`1`),
			},
			want: []Diff{
				{Stack: []string{}, Src: 0., Dist: 1.},
			},
		},
		{
			args: args{
				djson(`[1,2]`),
				djson(`[1,2]`),
			},
		},
		{
			args: args{
				djson(`[1]`),
				djson(`[1,2]`),
			},
			want: []Diff{{Stack: []string{}, Src: []interface{}{1.}, Dist: []interface{}{1., 2.}}},
		},
		{
			args: args{
				djson(`[1,2]`),
				djson(`[1]`),
			},
			want: []Diff{{Stack: []string{}, Src: []interface{}{1., 2.}, Dist: []interface{}{1.}}},
		},
		{
			args: args{
				djson(`{"a":"z"}`),
				djson(`{"a":"z"}`),
			},
		},
		{
			args: args{
				djson(`{"a":"z"}`),
				djson(`{"a":"z","b":"y"}`),
			},
			want: []Diff{{Stack: []string{"b"}, Src: nil, Dist: "y"}},
		},
		{
			args: args{
				djson(`{"a":"z","b":"y"}`),
				djson(`{"a":"z"}`),
			},
			want: []Diff{{Stack: []string{"b"}, Src: "y", Dist: nil}},
		},

		{
			args: args{
				djson(`{"v":[1,2]}`),
				djson(`{"v":[1,2]}`),
			},
		},
		{
			args: args{
				djson(`{"v":[1]}`),
				djson(`{"v":[1,2]}`),
			},
			want: []Diff{{Stack: []string{"v"}, Src: []interface{}{1.}, Dist: []interface{}{1., 2.}}},
		},
		{
			args: args{
				djson(`{"v":[1,2]}`),
				djson(`{"v":[1]}`),
			},
			want: []Diff{{Stack: []string{"v"}, Src: []interface{}{1., 2.}, Dist: []interface{}{1.}}},
		},
		{
			args: args{
				djson(`{"v":{"a":"z"}}`),
				djson(`{"v":{"a":"z"}}`),
			},
		},
		{
			args: args{
				djson(`{"v":{"a":"z"}}`),
				djson(`{"v":{"a":"z","b":"y"}}`),
			},
			want: []Diff{{Stack: []string{"v", "b"}, Src: nil, Dist: "y"}},
		},
		{
			args: args{
				djson(`{"v":{"a":"z","b":"y"}}`),
				djson(`{"v":{"a":"z"}}`),
			},
			want: []Diff{{Stack: []string{"v", "b"}, Src: "y", Dist: nil}},
		},

		{
			args: args{
				djson(`[[1,2]]`),
				djson(`[[1,2]]`),
			},
		},
		{
			args: args{
				djson(`[[1]]`),
				djson(`[[1,2]]`),
			},
			want: []Diff{{Stack: []string{"0"}, Src: []interface{}{1.}, Dist: []interface{}{1., 2.}}},
		},
		{
			args: args{
				djson(`[[1,2]]`),
				djson(`[[1]]`),
			},
			want: []Diff{{Stack: []string{"0"}, Src: []interface{}{1., 2.}, Dist: []interface{}{1.}}},
		},
		{
			args: args{
				djson(`[{"a":"z"}]`),
				djson(`[{"a":"z"}]`),
			},
		},
		{
			args: args{
				djson(`[{"a":"z"}]`),
				djson(`[{"a":"z","b":"y"}]`),
			},
			want: []Diff{{Stack: []string{"0", "b"}, Src: nil, Dist: "y"}},
		},
		{
			args: args{
				djson(`[{"a":"z","b":"y"}]`),
				djson(`[{"a":"z"}]`),
			},
			want: []Diff{{Stack: []string{"0", "b"}, Src: "y", Dist: nil}},
		},

		{
			args: args{
				djson(`{"v":{"a":"z","b":[1,2]}}`),
				djson(`{"v":{"a":"z","b":[1]}}`),
			},
			want: []Diff{{Stack: []string{"v", "b"}, Src: []interface{}{1., 2.}, Dist: []interface{}{1.}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeepDiff(tt.args.x, tt.args.y); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeepDiff() = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func djson(s string) interface{} {
	var i interface{}
	err := json.Unmarshal([]byte(s), &i)
	if err != nil {
		err = fmt.Errorf("Unmarshal json %q: error %w", s, err)
		panic(err)
	}
	return i
}
