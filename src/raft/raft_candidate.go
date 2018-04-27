package raft

func (rf *Raft) contestAnElection() {
	debugPrintf("[server:%v]state:%s\n, 开始给自己拉票精选", rf.me, rf.state)

	// increment currentTerm
	// TODO: 为什么不在这里 ++
	//rf.currentTerm++
	// vote for itself

	// 先给自己投一票
	rf.votedFor = rf.me
	// 现在总的投票人数为 1，就是自己投给自己的那一票
	rf.votesForMe = 1

	// 根据自己的参数，生成新的 requestVoteArgs
	// 发给所有人的都是一样的，所以只用生成一份
	requestVoteArgs := rf.newRequestVoteArgs()

	// 通过 requestVoteReplyChan 获取 goroutine 获取的 reply
	requestVoteReplyChan := make(chan *RequestVoteReply)

	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}

		go func(server int, args *RequestVoteArgs, replyChan chan *RequestVoteReply) {
			// ok := rf.sendRequestVote(server, args, reply)
			reply := new(RequestVoteReply)
			rf.sendRequestVote(server, args, reply)
			rf.mu.Lock()
			if rf.state != CANDIDATE {
				rf.mu.Unlock()
				return
			}
			rf.mu.Unlock()

			replyChan <- reply

		}(server, requestVoteArgs, requestVoteReplyChan)
	}

	reply := new(RequestVoteReply)
	totalReturns := 0
loop:
	for {
		select {
		// election timout elapses: start new election
		case <-rf.t.C:
			//rf.t.Stop()
			rf.timerReset()
			break loop
		case reply = <-requestVoteReplyChan:
			totalReturns++
			if !reply.VoteGranted {
				continue
			}
			rf.votesForMe++
			if rf.votesForMe > len(rf.peers)/2 {
				rf.comeToPower()
				break loop
			}
		default:
			// TODO: 为什么会有这个呢：
			rf.mu.Unlock()

			rf.mu.Lock()
			if rf.state == FOLLOWER {
				break loop
			}
		}
	}

	debugPrintf("[server: %v]Total granted peers: %v, total peers: %v\n", rf.me, rf.votesForMe, len(rf.peers))
	rf.mu.Unlock()
}

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
