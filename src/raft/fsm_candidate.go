package raft

var (
	winThisTermElectionEvent   = fsmEvent("win this term election")
	winThisTermElectionHandler = fsmHandler(
		func(rf *Raft) fsmState {

			return LEADER
		})
)

// 添加 CANDIDATE 状态下的处理函数
func (rf *Raft) addCandidateHandler() {
	rf.addHandler(CANDIDATE, electionTimeOutEvent, electionTimeOutHandler)
}
