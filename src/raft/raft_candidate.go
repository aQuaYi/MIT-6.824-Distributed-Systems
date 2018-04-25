package raft

// rf 为自己拉票，以便赢得选举
func (rf *Raft) canvass() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateID:  rf.me,
		LastLogTerm:  rf.getLastTerm(),
		LastLogIndex: rf.getLastIndex(),
	}

	for i := range rf.peers {
		if i != rf.me && rf.state == CANDIDATE {
			go func(i int) {
				var reply RequestVoteReply
				rf.sendRequestVote(i, &args, &reply)

				// NOTICE: 后续如何处理
			}(i)
		}
	}

	return
}
