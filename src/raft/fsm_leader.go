package raft

var (
	meetHigherTermLeaderEvent = fsmEvent("meet a leader with higher term")
)

// 添加 LEADER 状态下的处理函数
func (rf *Raft) addLeaderHandler() {
	rf.addHandler(LEADER, meetHigherTermLeaderEvent, fsmHandler(fallToFollower))
}

func fallToFollower(rf *Raft) fsmState {
	return FOLLOWER
}
