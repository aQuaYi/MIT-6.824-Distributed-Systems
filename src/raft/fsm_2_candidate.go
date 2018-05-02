package raft

import (
	"time"
)

// 引用时 args 为 nil
func comeToPower(rf *Raft, args interface{}) fsmState {

	debugPrintf("# %s # come to power", rf)

	// 新当选的 Leader 需要重置以下两个属性
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = len(rf.logs)
		rf.matchIndex[i] = 0
	}

	go sendHeartbeat(rf)

	return LEADER
}

// TODO: finish this goroutine
func sendHeartbeat(rf *Raft) {
	hbPeriod := time.Duration(100) * time.Millisecond
	hbtimer := time.NewTicker(hbPeriod)

	debugPrintf("# %s # 准备开始发送周期性心跳，周期:%s", rf, hbPeriod)

	for {
		// 先把自己的 timer 重置了，免得自己又开始新的 election
		rf.electionTimerReset()

		// TODO: 并行地给 所有的 FOLLOWER 发送 appendEntries RPC
		go makeHeartbeat(rf)

		select {
		// 要么 leader 变成了 follower，就只能结束这个循环
		case <-rf.convertToFollowerChan:
			return
		// 要么此次 heartbeat 结束
		case <-hbtimer.C:
		}
	}
}

type followToArgs struct {
	term     int
	votedFor int
}

// 发现了 leader with new term
func followTo(rf *Raft, args interface{}) fsmState {
	a, ok := args.(followToArgs)
	if !ok {
		panic("followTo 需要正确的参数")
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

func makeHeartbeat(rf *Raft) {
	for server := range rf.peers {
		if server == rf.me {
			continue
		}
		// appendEntriesArgs[server] = newAppendEntriesArgs(rf, server)
		// appendEntriesReply[server] = new(AppendEntriesReply)

		go func(server int) {
			args, reply := newAppendEntriesArgs(rf, server), new(AppendEntriesReply)
			ok := rf.sendAppendEntries(server, args, reply)

			if !ok {
				return
			}

			rf.rwmu.Lock()
			defer rf.rwmu.Unlock()

			// if last log index >= nextIndex for a follower:
			// send AppendEntries RPC with log entries starting at nextIndex
			// 1) if successful: update nextIndex and matchIndex for follower
			// 2) if AppendEntries fails because of log inconsistency:
			//    decrement nextIndex and retry

			// if get an old RPC reply
			// TODO: 为什么接收到一个 old rpc reply

			if args.Term != rf.currentTerm {
				// rf.rwmu.Unlock()
				return
			}

			// check matchIndex in case RPC is lost during reply of log recovery
			// log is attached to follower but leader not receive success reply
			if reply.Success {
				if rf.matchIndex[server] < args.PrevLogIndex {
					rf.matchIndex[server] = args.PrevLogIndex
					rf.appendedNewEntriesChan <- struct{}{}
				}
				return
			}

			if rf.state != LEADER {
				// rf.rwmu.Unlock()
				return
			}

			if reply.Term > rf.currentTerm {
				go rf.call(discoverNewTermEvent, followToArgs{
					term:     reply.Term,
					votedFor: NULL,
				})
				return
			}

			var firstTermIndex int
			rf.nextIndex[server] = reply.NextIndex
			for {
				debugPrintf("abc:%v, server: %v reply: %v\n", rf, server, reply)
				detectAppendEntriesArgs := newAppendEntriesArgs(rf, server)
				detectReply := new(AppendEntriesReply)

				rf.rwmu.Unlock()
				ok := rf.sendAppendEntries(server, detectAppendEntriesArgs, detectReply)
				rf.rwmu.Lock()

				if !ok {
					debugPrintf("[server: %v]not receive from %v\n", rf.me, server)
					rf.nextIndex[server] = len(rf.logs)
					return
				}

				if args.Term != rf.currentTerm {
					return
				}

				if detectReply.Term > rf.currentTerm {
					go rf.call(discoverNewTermEvent, followToArgs{
						term:     reply.Term,
						votedFor: NULL,
					})
					return
				}

				if detectReply.Success {
					firstTermIndex = detectAppendEntriesArgs.PrevLogIndex + 1
					break
				}

				rf.nextIndex[server] = detectReply.NextIndex
			}

			debugPrintf("[server: %v]Consistency check: server: %v, firstTermIndex: %v", rf.me, server, firstTermIndex)
			forceAppendEntriesArgs := newForceAppendEntriesArgs(rf, firstTermIndex)
			forceReply := new(AppendEntriesReply)

			rf.rwmu.Unlock()
			ok = rf.sendAppendEntries(server, forceAppendEntriesArgs, forceReply)
			rf.rwmu.Lock()

			if !ok {
				debugPrintf("[server: %v]no reponse from %v\n", rf.me, server)
				rf.nextIndex[server] = len(rf.logs)
				return
			}

			if args.Term != rf.currentTerm {
				// rf.rwmu.Unlock()
				return
			}

			if forceReply.Term > rf.currentTerm {
				go rf.call(discoverNewTermEvent, followToArgs{
					term:     reply.Term,
					votedFor: NULL,
				})
				return
			}

			debugPrintf("[server: %v]successfully append entries: %v\n", rf.me, forceReply)
			rf.nextIndex[server] = len(rf.logs)
			rf.matchIndex[server] = forceAppendEntriesArgs.PrevLogIndex + len(forceAppendEntriesArgs.Entries)
			// rf.rwmu.Unlock()
			rf.appendedNewEntriesChan <- struct{}{}
			return

			// rf.rwmu.Unlock()
		}(server)
	}

	return
}
