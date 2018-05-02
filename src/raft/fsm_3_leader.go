package raft

var (
	discoverHigherTermLeaderEvent = fsmEvent("meet a leader with higher term")
)

// 添加 LEADER 状态下的处理函数
func (rf *Raft) addLeaderHandler() {
	rf.addHandler(LEADER, discoverHigherTermLeaderEvent, fsmHandler(convertToFollower))
}
