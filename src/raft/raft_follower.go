package raft

import "time"

func (rf *Raft) standingBy() {
	select {
	case <-rf.t.C:
		debugPrintf("[server: %v]change to candidate\n", rf.me)
		rf.state = CANDIDATE
		// reset election timer
		//
		// TODO: 删除此处内容
		// rf.t.Reset(timeout * time.Millisecond)
		//
		rf.timerReset()
	default:
	}
	rf.mu.Unlock()
	time.Sleep(1 * time.Millisecond)
}
