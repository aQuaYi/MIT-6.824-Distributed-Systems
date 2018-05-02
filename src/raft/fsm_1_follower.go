package raft

// election time out 意味着，
// 进入新的 term
// 并开始新一轮的选举
func startNewElection(rf *Raft, null interface{}) fsmState {

	// 先进入下一个 Term
	rf.currentTerm++

	if rf.electionTimeoutChan != nil {
		close(rf.electionTimeoutChan)
	}
	rf.electionTimeoutChan = make(chan struct{})

	// 先给自己投一票
	rf.votedFor = rf.me

	rf.convertToFollowerChan = make(chan struct{})

	// 通过 requestVoteReplyChan 获取 goroutine 获取的 reply
	requestVoteReplyChan := make(chan *RequestVoteReply, len(rf.peers))
	// 向每个 server 拉票

	debugPrintf("# %s #  在 term(%d) 开始拉票", rf, rf.currentTerm)

	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}
		go func(server int, replyChan chan *RequestVoteReply) {
			args := rf.newRequestVoteArgs()
			// 生成投票结果变量
			reply := new(RequestVoteReply)
			// 拉票
			rf.sendRequestVote(server, args, reply)
			// 返回投票结果
			replyChan <- reply
		}(server, requestVoteReplyChan)
	}

	go func(replyChan chan *RequestVoteReply) {
		convertToFollowerChan := rf.convertToFollowerChan
		// 现在总的投票人数为 1，就是自己投给自己的那一票
		votesForMe := 1
		for {
			select {
			case <-convertToFollowerChan:
				// rf 不再是 candidate 状态
				// 没有必要再统计投票结果了
				return
			case <-rf.electionTimeoutChan:
				// 新的 election 已经开始，可以结束这个了
				return
			case reply := <-requestVoteReplyChan: // 收到新的选票
				if reply.Term > rf.currentTerm {
					rf.call(discoverNewTermEvent,
						followToArgs{
							term:     reply.Term,
							votedFor: NULL,
						})
					// TODO: 这个 return 应该是 写不写都行
					return
				}
				if reply.IsVoteGranted {
					// 投票给我的人数 +1
					votesForMe++
					// 如果投票任务过半，那我就是新的 LEADER 了
					if votesForMe > len(rf.peers)/2 {
						rf.call(winThisTermElectionEvent, nil)
						return
					}
				}
			}
		}
	}(requestVoteReplyChan)

	return CANDIDATE
}

func convertToFollower(rf *Raft, term interface{}) fsmState {
	newTerm, ok := term.(int)
	if !ok {
		panic("convertToFollower 需要正确的参数")
	}
	rf.currentTerm = max(rf.currentTerm, newTerm)
	rf.votedFor = NULL

	// rf.convertToFollowerChan != nil
	// 说明，rf 是 candidate 或 leader
	if rf.convertToFollowerChan != nil {
		close(rf.convertToFollowerChan)
		rf.convertToFollowerChan = nil
	}

	// TODO: 这里需要重置 election timer 吗

	return FOLLOWER
}
