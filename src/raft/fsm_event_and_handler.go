package raft

// for all servers
var (
	//
	electionTimeOutEvent          = fsmEvent("election time out")
	discoverNewTermEvent          = fsmEvent("discover new term")
	winThisTermElectionEvent      = fsmEvent("win this term election")
	discoverCurrentLeaderEvent    = fsmEvent("discovers current leader")
	discoverHigherTermLeaderEvent = fsmEvent("meet a leader with higher term")
)

// 添加 FOLLOWER 状态下的处理函数
func (rf *Raft) addFollowerHandler() {
	rf.addHandler(FOLLOWER, electionTimeOutEvent, fsmHandler(startNewElection))
	rf.addHandler(FOLLOWER, discoverNewTermEvent, fsmHandler(convertToFollower))
	rf.addHandler(FOLLOWER, discoverHigherTermLeaderEvent, fsmHandler(followTo))
}

// 添加 CANDIDATE 状态下的处理函数
func (rf *Raft) addCandidateHandler() {
	rf.addHandler(CANDIDATE, winThisTermElectionEvent, fsmHandler(comeToPower))
	rf.addHandler(CANDIDATE, discoverCurrentLeaderEvent, fsmHandler(followTo))
	rf.addHandler(CANDIDATE, discoverNewTermEvent, fsmHandler(convertToFollower))
	rf.addHandler(CANDIDATE, electionTimeOutEvent, fsmHandler(startNewElection))
}

// 添加 LEADER 状态下的处理函数
func (rf *Raft) addLeaderHandler() {
	rf.addHandler(LEADER, discoverHigherTermLeaderEvent, fsmHandler(convertToFollower))
}

func (rf *Raft) addAllHandler() {
	rf.addFollowerHandler()
	rf.addCandidateHandler()
	rf.addLeaderHandler()
}
