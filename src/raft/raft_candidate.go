package raft

import "time"

// 发起一场竞选
func (rf *Raft) contestAnElection() {
	debugPrintf("[server: %v]state:%s, 开始给自己拉票竞选", rf.me, rf.state)

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

	// 向每个 server 拉票
	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}

		go func(server int, args *RequestVoteArgs, replyChan chan *RequestVoteReply) {
			// 生成投票结果变量
			reply := new(RequestVoteReply)
			// 拉票
			rf.sendRequestVote(server, args, reply)

			rf.rwmu.Lock()
			// 如果 rf 已经不是 CANDIDATE 了
			// 可以提前结束，不用反馈投票结果
			if rf.state != CANDIDATE {
				rf.rwmu.Unlock()
				return
			}
			rf.rwmu.Unlock()

			// 返回投票结果
			replyChan <- reply

		}(server, requestVoteArgs, requestVoteReplyChan)
	}

	reply := new(RequestVoteReply)
loop:
	for {
		select {
		// 选举时间结束，需要开始新的选举
		case <-rf.electionTimer.C:
			//rf.t.Stop()
			rf.electionTimerReset()
			// 结束 loop 循环
			break loop
			// 收到新的选票
		case reply = <-requestVoteReplyChan:
			// 不是投票给我的，就不用继续了
			if !reply.VoteGranted {
				continue
			}

			// 投票给我的人数 +1
			rf.votesForMe++
			// 如果投票任务过半，那我就是新的 LEADER 了
			if rf.votesForMe > len(rf.peers)/2 {
				rf.comeToPower()
				break loop
			}
		default:
			// TODO: 为什么会有这个呢：
			rf.rwmu.Unlock()
			time.Sleep(1 * time.Millisecond)
			rf.rwmu.Lock()
			if rf.state == FOLLOWER {
				// rf 已经由于其他原因，转换成 FOLLOWER 了
				// 也可以直接结束 loop 循环了
				break loop
			}
		}
	}

	debugPrintf("[server: %v]Total granted peers: %v, total peers: %v\n", rf.me, rf.votesForMe, len(rf.peers))
	rf.rwmu.Unlock()
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

// TODO: 如果有 5 个 server，1,2 投票给 3， 3 成为 leader， 4 投票给 5， 5 没有赢得选举，但是，5 会 5.currentTerm++ 然后，开始新的 election，由于 只有 5.currentTerm 增加了，5 会成为新的 Leader。怎么破？
