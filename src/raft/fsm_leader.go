package raft

var (
	meetHigherTermLeaderEvent   = fsmEvent("meet a leader with higher term")
	meetHigherTermLeaderHandler = fsmHandler(
		func(rf *Raft) fsmState {
			return FOLLOWER
		})
)

// 添加 LEADER 状态下的处理函数
func (rf *Raft) addLeaderHandler() {
	rf.addHandler(LEADER, meetHigherTermLeaderEvent, meetHigherTermLeaderHandler)
}
