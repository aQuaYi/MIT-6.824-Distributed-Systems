package raft

// ApplyMsg 是发送消息
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct { // TODO: 注释 applyMsg 中每个属性的含义
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

// TODO: 这个函数是干什么用的
func (rf *Raft) reportApplyMsg(applyCh chan ApplyMsg) {
	for {
		if rf.hasShutdown() {
			debugPrintf("[server: %v]Close logs handling goroutine\n", rf.me)
			//rf.mu.Unlock()
			return
		}

		<-rf.appendedNewEntriesChan

		matchIndexCntr := make(map[int]int)
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

		if rf.lastApplied == rf.commitIndex {
			continue
		}

		debugPrintf("[server: %v]lastApplied: %v, commitIndex: %v\n", rf.me, rf.lastApplied, rf.commitIndex)
		for rf.lastApplied < rf.commitIndex {
			rf.lastApplied++
			applyMsg := ApplyMsg{
				CommandValid: true,
				Command:      rf.logs[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied}
			debugPrintf("[server: %v]send committed log to service: %v\n", rf.me, applyMsg)
			// rf.mu.Unlock()
			applyCh <- applyMsg
			// rf.mu.Lock()
		}

		// persist only when possible committed data
		// for leader, it's easy to determine
		// persist leader during commit
		if rf.state == LEADER {
			rf.persist()
		}

	}
}
