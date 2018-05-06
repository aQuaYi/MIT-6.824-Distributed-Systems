package raft

import (
	"time"
)

// election time out 意味着，
// 进入新的 term
// 并开始新一轮的选举
func startNewElection(rf *Raft, null interface{}) fsmState {
	// 先进入下一个 Term
	rf.currentTerm++

	// 如果 rf 转换前的状态是 Candidate，
	// 前一个 election 还没有结束, 需要通知前一个 election 彻底关闭
	if rf.endElectionChan != nil {
		close(rf.endElectionChan)
	}
	rf.endElectionChan = make(chan struct{})

	// 先给自己投一票
	rf.votedFor = rf.me

	// 如果在此次 election 中，别的 server 当选
	// 此 server 会转变成 FOLLOWER 状态
	// 需要关闭这个 channel 来发送通知
	rf.convertToFollowerChan = make(chan struct{})

	// 通过 replyChan 发送获取的 VoteReply 到同一个 goroutine 进行统计
	replyChan := make(chan *RequestVoteReply, len(rf.peers))

	debugPrintf("%s 在 term(%d) 开始拉票", rf, rf.currentTerm)

	// 向每个 server 拉票
	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}

		// 并行拉票
		go func(server int, replyChan chan *RequestVoteReply) {
			args := rf.newRequestVoteArgs()
			// 生成投票结果变量
			reply := new(RequestVoteReply)

			// 拉票
			ok := rf.sendRequestVote(server, args, reply)
			if !ok {
				debugPrintf("%s 无法获取 S#%d 对 %s 的反馈", rf, server, args)
				return
			}

			if args.Term == rf.currentTerm && rf.state == CANDIDATE {
				// 返回投票结果
				replyChan <- reply
				debugPrintf("%s 已经收集 S#%d 对 %s 的 %s", rf, server, args, reply)
			}

		}(server, replyChan)
	}

	go func(replyChan chan *RequestVoteReply) {
		// 现在总的投票人数为 1，就是自己投给自己的那一票
		votesForMe := 1
		debugPrintf("%s 等待 term(%d) 的选票", rf, rf.currentTerm)

		for {

			select {
			case <-rf.convertToFollowerChan:
				// rf 不再是 candidate 状态
				// 没有必要再统计投票结果了
				debugPrintf("%s 已经是 %s，停止统计投票的工作", rf, rf.state)
				return
			case <-rf.endElectionChan:
				// 新的 election 已经开始，可以结束这个了
				debugPrintf("%s 收到 election timeout 的信号，停止统计投票的工作", rf, rf.state)
				return
			case reply := <-replyChan: // 收到新的选票
				//
				if reply.Term > rf.currentTerm {
					rf.call(discoverNewTermEvent,
						toFollowerArgs{
							term:     reply.Term,
							votedFor: NOBODY,
						})
					return
				}
				if reply.IsVoteGranted {
					// 投票给我的人数 +1
					votesForMe++
					// 如果投票任务过半，那我就是新的 LEADER 了
					if votesForMe > len(rf.peers)/2 {
						rf.call(winElectionEvent, nil)
						return
					}
				}
			}
		}
	}(replyChan)

	return CANDIDATE
}

// 引用时 args 为 nil
func comeToPower(rf *Raft, args interface{}) fsmState {
	debugPrintf("%s come to power", rf)

	// 及时清理 channel
	if rf.endElectionChan != nil {
		close(rf.endElectionChan)
		rf.endElectionChan = nil
	}

	// 及时清理 channel
	if rf.convertToFollowerChan != nil {
		close(rf.convertToFollowerChan)
	}
	rf.convertToFollowerChan = make(chan struct{})

	// 新当选的 Leader 需要重置以下两个属性
	rf.nextIndex = make([]int, len(rf.peers))
	for i := range rf.nextIndex {
		rf.nextIndex[i] = len(rf.logs)
	}
	rf.matchIndex = make([]int, len(rf.peers))

	go heartbeating(rf)

	return LEADER
}

func heartbeating(rf *Raft) {
	hbPeriod := time.Duration(100) * time.Millisecond
	hbTimer := time.NewTicker(hbPeriod)

	debugPrintf("%s 开始发送周期性心跳，周期:%s", rf, hbPeriod)

	for {
		// 对于自己只用直接重置 timer
		rf.resetElectionTimerChan <- struct{}{}

		// 并行地给 所有的 FOLLOWER 发送 appendEntries RPC
		go oneHearteat(rf)

		select {
		// 要么，leader 变成了 follower，就只能结束这个循环
		case <-rf.convertToFollowerChan:
			return
		// 要么，开始下一次循环
		case <-hbTimer.C:
		}
	}
}

func oneHearteat(rf *Raft) {

	// 如果，在发送 RPC 前才生成 args 的话
	//
	// 某个旧 term 的 oldLeader 掉线很久以后，还是会一直坚持不懈地发送 appendEntries RPC
	// oldLeader 在其最后一个 oneHeartbeat 刚刚开始后，重新上线后
	// 在收到此 oneHeartbeat 的第一个回复后，
	// 会发现自己的 Term 落后了，然后更新
	// 		oldLeader.currentTerm 为最新值
	// 		oldLeader.state 为 FOLLOWER
	// ** 但是 **
	// 此 oneHeartbeat 还没有发送的 RPC 还会继续发送
	// rpc args.Term 却是更新后的最新 term 了
	// 此时就相当于，此 term 有两个 leader 在向外部发送 appendEntries RPC
	//
	// 所以，在 oneHeartbeat 的开始，就一口气生成后所有的 args
	// 然后，如果 rf 不再是 leader 的话，就取不再发送
	// 这样的话，即使 rf 在 oneHeartbeat 内不再是 leader
	// 其 rpc args 的内容，还是以前的内容，不会出现，同一个 term 两个 leader 的情况
	argsSlice := genAppendEntriesArgsSlice(rf)
	if rf.state != LEADER {
		return
	}

	for server := range rf.peers {
		if server == rf.me {
			continue
		}

		go func(server int, args *AppendEntriesArgs) {
			reply := new(AppendEntriesReply)

			ok := rf.sendAppendEntries(server, args, reply)

			if !ok {
				debugPrintf("%s 无法获取 S#%d 对 %s 的回复", rf, server, args)
				return
			}

			if reply.Term > rf.currentTerm {
				go rf.call(discoverNewTermEvent, toFollowerArgs{
					term:     reply.Term,
					votedFor: NOBODY,
				})
				return
			}

			// if get an old RPC reply
			// TODO: 为什么接收到一个 old rpc reply
			if args.Term != rf.currentTerm {
				return
			}

			rf.rwmu.Lock()
			defer rf.rwmu.Unlock()

			// if last log index >= nextIndex for a follower:
			// send AppendEntries RPC with log entries starting at nextIndex
			// 1) if successful: update nextIndex and matchIndex for follower
			// 2) if AppendEntries fails because of log inconsistency:
			//    decrement nextIndex and retry

			if reply.Success {
				// reply.Success == true，表示，args.Entries 确实添加到了 server.logs
				// 但是，由于每 100ms 就会发送一个 appendEntries RPC
				// 所以，很有可能，先发送 RPC 返回的 reply 会延后几个周期才被处理
				// 此时 reply.NextIndex 的确有可能比 rf.nextIndex[server] 小
				// 所以，使用 max 来更新
				rf.nextIndex[server] = max(rf.nextIndex[server], reply.NextIndex)
				if rf.matchIndex[server] < rf.nextIndex[server]-1 {
					rf.matchIndex[server] = rf.nextIndex[server] - 1
					rf.toCheckApplyChan <- struct{}{}
				}
			} else {
				// 使用 min 来更新，与 reply.Success == true 同理
				rf.nextIndex[server] = min(rf.nextIndex[server], reply.NextIndex)
			}

		}(server, argsSlice[server])
	}

	return
}

func genAppendEntriesArgsSlice(rf *Raft) []*AppendEntriesArgs {
	res := make([]*AppendEntriesArgs, len(rf.peers))
	for server := range rf.peers {
		if server == rf.me {
			continue
		}
		res[server] = newAppendEntriesArgs(rf, server)
	}
	return res
}

type toFollowerArgs struct {
	term     int
	votedFor int
}

// dicover leader or new term
func toFollower(rf *Raft, args interface{}) fsmState {
	a, ok := args.(toFollowerArgs)
	if !ok {
		panic("toFollower 需要正确的参数")
	}

	// 遇到新 term 的话，需要更新 currentTerm
	rf.currentTerm = max(rf.currentTerm, a.term)
	// 遇到新 leader 的话，需要更新 votedFor
	rf.votedFor = a.votedFor

	// rf.convertToFollowerChan != nil 就一定是 open 的
	// 这是靠锁保证的
	if rf.convertToFollowerChan != nil {
		close(rf.convertToFollowerChan)
		rf.convertToFollowerChan = nil
	}

	if rf.endElectionChan != nil {
		close(rf.endElectionChan)
		rf.endElectionChan = nil
	}

	return FOLLOWER
}
