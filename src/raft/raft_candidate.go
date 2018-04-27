package raft

import "time"

func (rf *Raft) contestAnElection() {
	debugPrintf("[server:%v]state:%s\n, 开始给自己拉票精选", rf.me, rf.state)

	// increment currentTerm
	// TODO: 为什么不在这里 ++
	//rf.currentTerm++
	// vote for itself

	// 先给自己投一票
	rf.votedFor = rf.me
	grantedCnt := 1

	requestVoteArgs := newRequestVoteArgs(rf)

	// send RequestVote to all other servers
	requestVoteReplyChan := make(chan *RequestVoteReply)

	requestVoteReply := make([]*RequestVoteReply, len(rf.peers))

	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}

		requestVoteReply[server] = new(RequestVoteReply)

		go func(server int, args *RequestVoteArgs, reply *RequestVoteReply, replyChan chan *RequestVoteReply) {
			// ok := rf.sendRequestVote(server, args, reply)
			rf.sendRequestVote(server, args, reply)
			rf.mu.Lock()
			if rf.state != CANDIDATE {
				rf.mu.Unlock()
				return
			}
			rf.mu.Unlock()

			replyChan <- reply

			// if ok && reply.VoteGranted {
			// 	replyChan <- reply
			// } else {
			// 	reply.VoteGranted = false
			// 	replyChan <- reply
			// }

		}(server, requestVoteArgs, requestVoteReply[server], requestVoteReplyChan)
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
			grantedCnt++
			if grantedCnt > len(rf.peers)/2 {
				rf.comeToPower()
				break loop
			}
		default:
			rf.mu.Unlock()
			time.Sleep(1 * time.Millisecond)
			rf.mu.Lock()
			if rf.state == FOLLOWER {
				break loop
			}
		}
	}

	debugPrintf("[server: %v]Total granted peers: %v, total peers: %v\n", rf.me, grantedCnt, len(rf.peers))
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
