package raft

import (
	"time"
)

var (
	winThisTermElectionEvent            = fsmEvent("win this term election")
	discoverCurrentLeaderOrNewTermEvent = fsmEvent("discovers current leader or new term")
)

// 添加 CANDIDATE 状态下的处理函数
func (rf *Raft) addCandidateHandler() {
	rf.addHandler(CANDIDATE, winThisTermElectionEvent, fsmHandler(comeToPower))
	rf.addHandler(CANDIDATE, discoverCurrentLeaderOrNewTermEvent, fsmHandler(candidateToFollower))
	rf.addHandler(CANDIDATE, electionTimeOutEvent, fsmHandler(startNewElection))

}

// 引用时 args 为 nil
func comeToPower(rf *Raft, args interface{}) fsmState {
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
		if !rf.isLeader() {
			return
		}

		// 先把自己的 timer 重置了，免得自己又开始新的 election
		rf.electionTimerReset()

		// TODO: 并行地给 所有的 FOLLOWER 发送 appendEntries RPC

		// 等待一段时间
		<-hbtimer.C
	}
}

// candidate to follower 有两个原因
// 	1. 发现了同 term 的leader
// 	2. 看见了更大的 term
// 所以，args 是一个 term，用于更新 rf.currentTerm
func candidateToFollower(rf *Raft, term interface{}) fsmState {
	newTerm, _ := term.(int)
	rf.currentTerm = max(rf.currentTerm, newTerm)

	// 这很重要
	rf.votedFor = NULL

	return FOLLOWER
}
