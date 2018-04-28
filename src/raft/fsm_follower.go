package raft

var (
	electionTimeOutEvent   = fsmEvent("election time out")
	electionTimeOutHandler = fsmHandler(
		func(rf *Raft) fsmState {
			return CANDIDATE
		})
)

// 添加 FOLLOWER 状态下的处理函数
func (rf *Raft) addFollowerHandler() {
	rf.addHandler(FOLLOWER, electionTimeOutEvent, electionTimeOutHandler)
}
