package raft

import "testing"

func Test_state_String(t *testing.T) {
	tests := []struct {
		name string
		s    fsmState
		want string
	}{

		{
			"print leader",
			LEADER,
			"Leader",
		},

		{
			"print candidate",
			CANDIDATE,
			"Candidate",
		},

		{
			"print follower",
			FOLLOWER,
			"Follower",
		},

		//
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("state.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
