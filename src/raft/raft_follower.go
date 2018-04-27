package raft

import "time"

func (rf *Raft) standingBy() {
	select {
	case <-rf.electionTimer.C:
		debugPrintf("[server: %v]change to candidate\n", rf.me)
		rf.state = CANDIDATE
		// reset election timer
		//
		// TODO: 需要在这里重置 timer 吗？
		rf.timerReset()
	default:
	}
	rf.mu.Unlock()

	time.Sleep(1 * time.Millisecond)
}
