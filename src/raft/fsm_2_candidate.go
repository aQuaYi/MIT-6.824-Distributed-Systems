package raft

import (
	"time"
)

// 引用时 args 为 nil
func comeToPower(rf *Raft, args interface{}) fsmState {
	rf.currentTerm++

	// 新当选的 Leader 需要重置以下两个属性
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = len(rf.logs)
		rf.matchIndex[i] = 0
	}

	go sendHeartbeat(rf)

	return LEADER
}

// TODO: finish this goroutine
func sendHeartbeat(rf *Raft) {
	hbPeriod := time.Duration(100)
	hbtimer := time.NewTicker(hbPeriod * time.Millisecond)

	for {
		// 先把自己的 timer 重置了，免得自己又开始新的 election
		rf.electionTimerReset()

		// TODO: 并行地给 所有的 FOLLOWER 发送 appendEntries RPC

		select {
		// 要么 leader 变成了 follower，就只能结束这个循环
		case <-rf.convertToFollowerChan:
			return
		// 要么此次 heartbeat 结束
		case <-hbtimer.C:
		}
	}
}

type followToArgs struct {
	term     int
	votedFor int
}

// 发现了 leader with new term
func followTo(rf *Raft, args interface{}) fsmState {
	a, ok := args.(followToArgs)
	if !ok {
		panic("followTo 需要正确的参数")
	}
	rf.currentTerm = max(rf.currentTerm, a.term)
	rf.votedFor = a.votedFor

	// rf.convertToFollowerChan != nil 就一定是 open 的
	// 这是靠锁保证的
	if rf.convertToFollowerChan != nil {
		close(rf.convertToFollowerChan)
		rf.convertToFollowerChan = nil
	}

	return FOLLOWER
}
