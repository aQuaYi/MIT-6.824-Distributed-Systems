package raft

import (
	"time"
)

// 引用时 args 为 nil
func comeToPower(rf *Raft, args interface{}) fsmState {
	debugPrintf("%s  come to power", rf)

	//
	if rf.electionTimeoutChan != nil {
		close(rf.electionTimeoutChan)
		rf.electionTimeoutChan = nil
	}

	// 新当选的 Leader 需要重置以下两个属性
	rf.nextIndex = make([]int, len(rf.peers))
	for i := range rf.nextIndex {
		rf.nextIndex[i] = len(rf.logs)
	}
	rf.matchIndex = make([]int, len(rf.peers))

	go sendHeartbeat(rf)

	return LEADER
}

// TODO: finish this goroutine
func sendHeartbeat(rf *Raft) {
	hbPeriod := time.Duration(100) * time.Millisecond
	hbtimer := time.NewTicker(hbPeriod)

	debugPrintf("%s  准备开始发送周期性心跳，周期:%s", rf, hbPeriod)

	for {
		// TODO: 并行地给 所有的 FOLLOWER 发送 appendEntries RPC
		makeHeartbeat(rf)

		select {
		// 要么 leader 变成了 follower，就只能结束这个循环
		case <-rf.convertToFollowerChan:
			return
		// 要么此次 heartbeat 结束
		case <-hbtimer.C:
		}
	}
}

func makeHeartbeat(rf *Raft) {
	for server := range rf.peers {
		if server == rf.me {
			// 对于自己只用直接重置 timer
			rf.electionTimerReset()
			continue
		}

		go func(server int) {
			args, reply := newAppendEntriesArgs(rf, server), new(AppendEntriesReply)
			ok := rf.sendAppendEntries(server, args, reply)

			if !ok {
				debugPrintf("%s  无法获取 S%d 对 %s 的回复", rf, server, args)
				return
			}

			if reply.Term > rf.currentTerm {
				go rf.call(discoverNewTermEvent, toFollowerArgs{
					term:     reply.Term,
					votedFor: NULL,
				})
				return
			}

			// if get an old RPC reply
			// TODO: 为什么接收到一个 old rpc reply
			if args.Term != rf.currentTerm {
				// rf.rwmu.Unlock()
				return
			}

			if rf.state != LEADER {
				// rf.rwmu.Unlock()
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
				rf.nextIndex[server] = max(rf.nextIndex[server], reply.NextIndex)
				rf.matchIndex[server] = rf.nextIndex[server] - 1
				rf.appendedNewEntriesChan <- struct{}{}
			} else {
				rf.nextIndex[server] = min(rf.nextIndex[server], reply.NextIndex)
			}

		}(server)
	}

	return
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

	rf.currentTerm = max(rf.currentTerm, a.term)
	rf.votedFor = a.votedFor

	// rf.convertToFollowerChan != nil 就一定是 open 的
	// 这是靠锁保证的
	if rf.convertToFollowerChan != nil {
		close(rf.convertToFollowerChan)
		rf.convertToFollowerChan = nil
	}

	return FOLLOWER
}
