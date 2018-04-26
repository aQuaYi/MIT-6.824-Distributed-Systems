package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"labrpc"
	"time"
)

const (
	// HBINTERVAL is haertbeat interval
	HBINTERVAL = 50 * time.Millisecond // 50ms

	// NULL 表示没有投票给任何人
	NULL = -1
)

// Make is
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := newRaft(peers, me, persister, applyCh)

	go func(rf *Raft) {
		for {
			// TODO: 为什么 for 循环一开始要 lock
			rf.mu.Lock()

			if rf.hasShutdown() {
				rf.mu.Unlock()
				return
			}

			switch rf.state {
			case FOLLOWER:
				select {
				case <-rf.t.C:
					debugPrintf("[server: %v]change to candidate\n", rf.me)
					rf.state = CANDIDATE
					// reset election timer
					//
					// TODO: 删除此处内容
					// rf.t.Reset(timeout * time.Millisecond)
					//
					rf.timerReset()
				default:
				}
				rf.mu.Unlock()
				time.Sleep(1 * time.Millisecond)

			case CANDIDATE:
				debugPrintf("[server:%v]state:%s\n", rf.me, rf.state)
				rf.contestAnElection()
			case LEADER:

				// Upon election: send initial hearbeat to each server
				// repeat during idle period to preven election timeout
				period := time.Duration(100)
				appendEntriesArgs := make([]*AppendEntriesArgs, len(peers))
				appendEntriesReply := make([]*AppendEntriesReply, len(peers))

				debugPrintf("[server: %v]Leader, send heartbeat, period: %v\n", rf.me, period*time.Millisecond)
				for server := range peers {
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
										if !rf.t.Stop() {
											debugPrintf("[server: %v]Leader change to follower1: drain timer\n", rf.me)
											<-rf.t.C
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
										if !rf.t.Stop() {
											debugPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
											<-rf.t.C
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
								if !rf.t.Stop() {
									debugPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
									<-rf.t.C
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
		}
	}(rf)

	go func(rf *Raft, applyCh chan ApplyMsg) {
		for {
			select {
			case <-rf.shutdown:
				debugPrintf("[server: %v]Close logs handling goroutine\n", rf.me)
				//rf.mu.Unlock()
				return
			default:
				matchIndexCntr := make(map[int]int)
				rf.mu.Lock()
				// update rf.commitIndex based on matchIndex[]
				// if there exists an N such that N > commitIndex, a majority of matchIndex[i] >= N
				// and log[N].term == currentTerm:
				// set commitIndex = N
				if rf.state == LEADER {
					rf.matchIndex[rf.me] = len(rf.logs) - 1
					for _, logIndex := range rf.matchIndex {
						if _, ok := matchIndexCntr[logIndex]; !ok {
							for _, logIndex2 := range rf.matchIndex {
								if logIndex <= logIndex2 {
									matchIndexCntr[logIndex]++
								}
							}
						}
					}
					// find the max matchIndex committed
					// paper 5.4.2, only log entries from the leader's current term are committed by counting replicas
					for index, matchNum := range matchIndexCntr {
						if matchNum > len(rf.peers)/2 && index > rf.commitIndex && rf.logs[index].LogTerm == rf.currentTerm {
							rf.commitIndex = index
						}
					}
					debugPrintf("[server: %v]matchIndex: %v, cntr: %v, rf.commitIndex: %v\n", rf.me, rf.matchIndex, matchIndexCntr, rf.commitIndex)
				}

				if rf.lastApplied < rf.commitIndex {
					debugPrintf("[server: %v]lastApplied: %v, commitIndex: %v\n", rf.me, rf.lastApplied, rf.commitIndex)
					for rf.lastApplied < rf.commitIndex {
						rf.lastApplied++
						applyMsg := ApplyMsg{
							CommandValid: true,
							Command:      rf.logs[rf.lastApplied].Command,
							CommandIndex: rf.lastApplied}
						debugPrintf("[server: %v]send committed log to service: %v\n", rf.me, applyMsg)
						rf.mu.Unlock()
						applyCh <- applyMsg
						rf.mu.Lock()
					}
					// persist only when possible committed data
					// for leader, it's easy to determine
					// persist leader during commit
					if rf.state == LEADER {
						rf.persist()
					}
				}
				rf.cond.Wait()
				rf.mu.Unlock()
			}
		}

	}(rf, applyCh)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
