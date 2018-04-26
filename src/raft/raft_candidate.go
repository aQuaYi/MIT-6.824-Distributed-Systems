package raft

func (rf *Raft) comeToPower() {
	// TODO: 这里有问题吧
	// 应该是先 Term++
	// 再进行选举
	rf.currentTerm++
	rf.state = LEADER

	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = len(rf.logs)
		rf.matchIndex[i] = 0
	}
}
