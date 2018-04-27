package raft

import "time"

func (rf *Raft) exercisePower() {
	// Upon election: send initial hearbeat to each server
	// repeat during idle period to preven election timeout
	period := time.Duration(100)
	appendEntriesArgs := make([]*AppendEntriesArgs, len(rf.peers))
	appendEntriesReply := make([]*AppendEntriesReply, len(rf.peers))

	debugPrintf("[server: %v]Leader, send heartbeat, period: %v\n", rf.me, period*time.Millisecond)
	for server := range rf.peers {
		if server == rf.me {
			continue
		}
		appendEntriesArgs[server] = newAppendEntriesArgs(rf, server)
		appendEntriesReply[server] = new(AppendEntriesReply)

		go func(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) {
			ok := rf.sendAppendEntries(server, args, reply)

			// if last log index >= nextIndex for a follower:
			// send AppendEntries RPC with log entries starting at nextIndex
			// 1) if successful: update nextIndex and matchIndex for follower
			// 2) if AppendEntries fails because of log inconsistency:
			//    decrement nextIndex and retry
			rf.mu.Lock()
			var firstTermIndex int
			// if get an old RPC reply
			if args.Term != rf.currentTerm {
				rf.mu.Unlock()
				return
			}

			// check matchIndex in case RPC is lost during reply of log recovery
			// log is attached to follower but leader not receive success reply
			if ok && reply.Success {
				if rf.matchIndex[server] < args.PrevLogIndex {
					rf.matchIndex[server] = args.PrevLogIndex
					rf.cond.Broadcast()
				}
			} else if ok && !reply.Success {
				if rf.state != LEADER {
					rf.mu.Unlock()
					return
				}
				if reply.Term <= rf.currentTerm {
					rf.nextIndex[server] = reply.FirstTermIndex
					for {
						//rf.nextIndex[server]--
						debugPrintf("abc:%v, server: %v reply: %v\n", rf, server, reply)
						detectAppendEntriesArgs := newAppendEntriesArgs(rf, server)
						rf.mu.Unlock()
						detectReply := new(AppendEntriesReply)
						ok1 := rf.sendAppendEntries(server, detectAppendEntriesArgs, detectReply)
						rf.mu.Lock()
						if !ok1 {
							debugPrintf("[server: %v]not receive from %v\n", rf.me, server)
							rf.nextIndex[server] = len(rf.logs)
							rf.mu.Unlock()
							return
						}
						if ok1 && args.Term != rf.currentTerm {
							rf.mu.Unlock()
							return
						}
						if detectReply.Term > rf.currentTerm {
							rf.state = FOLLOWER
							rf.currentTerm = detectReply.Term

							// reset timer
							if !rf.electionTimer.Stop() {
								debugPrintf("[server: %v]Leader change to follower1: drain timer\n", rf.me)
								<-rf.electionTimer.C
							}

							rf.timerReset()

							rf.mu.Unlock()
							return
						}
						if detectReply.Success {
							firstTermIndex = detectAppendEntriesArgs.PrevLogIndex + 1
							break
						}
						rf.nextIndex[server] = detectReply.FirstTermIndex
					}
					debugPrintf("[server: %v]Consistency check: server: %v, firstTermIndex: %v", rf.me, server, firstTermIndex)
					forceAppendEntriesArgs := newForceAppendEntriesArgs(rf, firstTermIndex)

					rf.mu.Unlock()
					forceReply := new(AppendEntriesReply)
					ok2 := rf.sendAppendEntries(server, forceAppendEntriesArgs, forceReply)
					rf.mu.Lock()
					if ok2 {
						if args.Term != rf.currentTerm {
							rf.mu.Unlock()
							return
						}
						if forceReply.Term > rf.currentTerm {
							rf.state = FOLLOWER
							rf.currentTerm = forceReply.Term

							// reset timer
							if !rf.electionTimer.Stop() {
								debugPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
								<-rf.electionTimer.C
							}

							rf.timerReset()

							rf.mu.Unlock()
							return
						}
						debugPrintf("[server: %v]successfully append entries: %v\n", rf.me, forceReply)
						rf.nextIndex[server] = len(rf.logs)
						rf.matchIndex[server] = forceAppendEntriesArgs.PrevLogIndex + len(forceAppendEntriesArgs.Entries)
						rf.mu.Unlock()
						rf.cond.Broadcast()
						return
					}
					debugPrintf("[server: %v]no reponse from %v\n", rf.me, server)
					rf.nextIndex[server] = len(rf.logs)
				} else {
					rf.state = FOLLOWER
					rf.currentTerm = reply.Term

					// reset timer
					if !rf.electionTimer.Stop() {
						debugPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
						<-rf.electionTimer.C
					}

					rf.timerReset()
				}
			}
			rf.mu.Unlock()
		}(server, appendEntriesArgs[server], appendEntriesReply[server])
	}

	rf.mu.Unlock()
	time.Sleep(period * time.Millisecond)
}
