package raft

var (
	electionTimeOutEvent = fsmEvent("election time out")
)

// 添加 FOLLOWER 状态下的处理函数
func (rf *Raft) addFollowerHandler() {
	rf.addHandler(FOLLOWER, electionTimeOutEvent, fsmHandler(startNewElection))
}

// election time out 意味着，
// 进入新的 term
// 并开始新一轮的选举
func startNewElection(rf *Raft, args interface{}) fsmState {
	// 先进入下一个 Term
	rf.currentTerm++
	// 先给自己投一票
	rf.votedFor = rf.me
	rf.votedTerm = rf.currentTerm

	rf.convertToFollowerChan = make(chan struct{})
	// TODO: 添加完需要修改的属性
	// TODO: 开个 goroutine 继续竞选
	return CANDIDATE
}
