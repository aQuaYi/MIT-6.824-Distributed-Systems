package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"labrpc"
)

const (
	// NULL 表示没有投票给任何人
	NULL = -1
)

// Make is
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//

// func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
// 	rf := newRaft(peers, me, persister)
// 	go func(rf *Raft) {
// 		for {
// 			// TODO: 为什么 for 循环一开始要 lock
// 			// 是为了等别的地方处理完，要不然直接进入
// 			// switch case FOLLOWER → select default → 就不会有超时什么事情了
// 			rf.rwmu.Lock()
// 			if rf.hasShutdown() {
// 				rf.rwmu.Unlock()
// 				return
// 			}
// 			switch rf.state {
// 			case FOLLOWER:
// 				rf.standingBy()
// 			case CANDIDATE:
// 				rf.contestAnElection()
// 			case LEADER:
// 				rf.exercisePower()
// 			}
// 		}
// 	}(rf)
// 	go reportApplyMsg(rf, applyCh)
// 	// initialize from state persisted before a crash
// 	rf.readPersist(persister.ReadRaftState())
// 	return rf
// }

// Make is
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := newRaft(peers, me, persister)

	go rf.reportApplyMsg(applyCh)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
