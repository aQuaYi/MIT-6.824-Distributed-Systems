package raft

import "testing"

func Test_maxMajorityIndex(t *testing.T) {
	type args struct {
		matchIndex []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{

		{
			"找到 max majority index 0",
			args{matchIndex: []int{1, 1, 0, 0, 0}},
			0,
		},

		{
			"找到 max majority index 6",
			args{matchIndex: []int{8, 7, 6, 5}},
			6,
		},

		{
			"找到 max majority index 6",
			args{matchIndex: []int{8, 7, 6, 5, 4}},
			6,
		},

		//
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxMajorityIndex(tt.args.matchIndex); got != tt.want {
				t.Errorf("maxMajorityIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
