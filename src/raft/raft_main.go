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
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := newRaft(peers, me, persister, applyCh)

	go func(rf *Raft) {
		for {
			// TODO: 为什么 for 循环一开始要 lock
			// 是为了等别的地方处理完，要不然直接进入
			// switch case FOLLOWER → select default → 就不会有超时什么事情了
			rf.mu.Lock()

			if rf.hasShutdown() {
				rf.mu.Unlock()
				return
			}

			switch rf.state {
			case FOLLOWER:
				rf.standingBy()
			case CANDIDATE:
				rf.contestAnElection()
			case LEADER:
				rf.exercisePower()
			}
		}
	}(rf)

	// TODO: 这个函数是干什么用的
	go func(rf *Raft, applyCh chan ApplyMsg) {
		for {
			if rf.hasShutdown() {
				debugPrintf("[server: %v]Close logs handling goroutine\n", rf.me)
				//rf.mu.Unlock()
				return
			}
			matchIndexCntr := make(map[int]int)
			rf.mu.Lock()
			// update rf.commitIndex based on matchIndex[]
			// if there exists an N such that N > commitIndex, a majority of matchIndex[i] >= N
			// and log[N].term == currentTerm:
			// set commitIndex = N
			if rf.state == LEADER {
				rf.matchIndex[rf.me] = len(rf.logs) - 1
				for _, logIndex := range rf.matchIndex {
					if _, ok := matchIndexCntr[logIndex]; !ok {
						for _, logIndex2 := range rf.matchIndex {
							if logIndex <= logIndex2 {
								matchIndexCntr[logIndex]++
							}
						}
					}
				}
				// find the max matchIndex committed
				// paper 5.4.2, only log entries from the leader's current term are committed by counting replicas
				for index, matchNum := range matchIndexCntr {
					if matchNum > len(rf.peers)/2 && index > rf.commitIndex && rf.logs[index].LogTerm == rf.currentTerm {
						rf.commitIndex = index
					}
				}
				debugPrintf("[server: %v]matchIndex: %v, cntr: %v, rf.commitIndex: %v\n", rf.me, rf.matchIndex, matchIndexCntr, rf.commitIndex)
			}

			if rf.lastApplied < rf.commitIndex {
				debugPrintf("[server: %v]lastApplied: %v, commitIndex: %v\n", rf.me, rf.lastApplied, rf.commitIndex)
				for rf.lastApplied < rf.commitIndex {
					rf.lastApplied++
					applyMsg := ApplyMsg{
						CommandValid: true,
						Command:      rf.logs[rf.lastApplied].Command,
						CommandIndex: rf.lastApplied}
					debugPrintf("[server: %v]send committed log to service: %v\n", rf.me, applyMsg)
					rf.mu.Unlock()
					applyCh <- applyMsg
					rf.mu.Lock()
				}
				// persist only when possible committed data
				// for leader, it's easy to determine
				// persist leader during commit
				if rf.state == LEADER {
					rf.persist()
				}
			}
			rf.cond.Wait()
			rf.mu.Unlock()
		}
	}(rf, applyCh)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
