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
	"math/rand"
	"time"
)

// GetState 可以获取 raft 对象的状态
// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// NOTICE: Your code here (2A).

	// TODO: 这里为什么要上锁呢
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term = rf.currentTerm
	isleader = rf.state == LEADER

	return term, isleader
}

func (rf *Raft) getLastIndex() int {
	return rf.logs[len(rf.logs)-1].LogIndex
}

func (rf *Raft) getLastTerm() int {
	return rf.logs[len(rf.logs)-1].LogTerm
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// NOTICE: Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// NOTICE: Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// Start 启动
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := rf.currentTerm
	isLeader := rf.state == LEADER
	if isLeader {
		index = rf.getLastIndex() + 1
		//fmt.Printf("raft:%d start\n",rf.me)
		rf.logs = append(rf.logs, LogEntry{LogTerm: term, LogCmd: command, LogIndex: index}) // append new entry from client
		rf.persist()
	}
	return index, term, isLeader
}

// Kill is
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// NOTICE: Your code here, if desired.
}

// rf 为自己拉票，以便赢得选举
func (rf *Raft) canvass() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateID:  rf.me,
		LastLogTerm:  rf.getLastTerm(),
		LastLogIndex: rf.getLastIndex(),
	}

	for i := range rf.peers {
		if i != rf.me && rf.state == CANDIDATE {
			go func(i int) {
				var reply RequestVoteReply
				rf.sendRequestVote(i, &args, &reply)

				// NOTICE: 后续如何处理
			}(i)
		}
	}

	return
}

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

	timeout := time.Duration(300 + rand.Int31n(400))
	rf.t = time.NewTimer(timeout * time.Millisecond)
	go func(rf *Raft) {
		for {
			rf.mu.Lock()
			if rf.state == CANDIDATE {
				DPrintf("[server: %v]state:%v\n", rf.me, rf.state)
			}
			select {
			case <-rf.shutdown:
				rf.mu.Unlock()
				DPrintf("[server: %v]Close state machine goroutine\n", rf.me)
				return
			default:
				switch rf.state {
				case FOLLOWER:
					select {
					case <-rf.t.C:
						DPrintf("[server: %v]change to candidate\n", rf.me)
						rf.state = CANDIDATE
						// reset election timer
						rf.t.Reset(timeout * time.Millisecond)
					default:
					}
					rf.mu.Unlock()
					time.Sleep(1 * time.Millisecond)

				case CANDIDATE:
					requestVoteArgs := new(RequestVoteArgs)
					requestVoteReply := make([]*RequestVoteReply, len(peers))

					// increment currentTerm
					//rf.currentTerm++
					// vote for itself
					rf.votedFor = rf.me
					grantedCnt := 1
					// send RequestVote to all other servers
					requestVoteArgs.Term = rf.currentTerm + 1
					requestVoteArgs.CandidateID = rf.me
					//requestVoteArgs.LastLogIndex = rf.commitIndex
					//requestVoteArgs.LastLogTerm  = rf.logs[rf.commitIndex].LogTerm
					requestVoteArgs.LastLogIndex = len(rf.logs) - 1
					requestVoteArgs.LastLogTerm = rf.logs[len(rf.logs)-1].LogTerm
					DPrintf("[server: %v] Candidate, election timeout %v, send RequestVote: %v\n", me, timeout*time.Millisecond, requestVoteArgs)

					requestVoteReplyChan := make(chan *RequestVoteReply)
					for server := range peers {
						if server != me {
							requestVoteReply[server] = new(RequestVoteReply)
							go func(server int, args *RequestVoteArgs, reply *RequestVoteReply, replyChan chan *RequestVoteReply) {
								ok := rf.sendRequestVote(server, args, reply)
								rf.mu.Lock()
								if rf.state != CANDIDATE {
									rf.mu.Unlock()
									return
								}
								rf.mu.Unlock()
								if ok && reply.VoteGranted {
									replyChan <- reply
								} else {
									reply.VoteGranted = false
									replyChan <- reply
								}
							}(server, requestVoteArgs, requestVoteReply[server], requestVoteReplyChan)
						}
					}

					reply := new(RequestVoteReply)
					totalReturns := 0
				loop:
					for {
						select {
						// election timout elapses: start new election
						case <-rf.t.C:
							//rf.t.Stop()
							timeout := time.Duration(500 + rand.Int31n(400))
							rf.t.Reset(timeout * time.Millisecond)
							break loop
						case reply = <-requestVoteReplyChan:
							totalReturns++
							if reply.VoteGranted {
								grantedCnt++
								if grantedCnt > len(peers)/2 {
									rf.currentTerm++
									rf.state = LEADER

									rf.nextIndex = make([]int, len(peers))
									rf.matchIndex = make([]int, len(peers))
									for i := 0; i < len(peers); i++ {
										rf.nextIndex[i] = len(rf.logs)
										rf.matchIndex[i] = 0
									}
									break loop
								}
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

					DPrintf("[server: %v]Total granted peers: %v, total peers: %v\n", rf.me, grantedCnt, len(peers))
					rf.mu.Unlock()

				case LEADER:

					// Upon election: send initial hearbeat to each server
					// repeat during idle period to preven election timeout
					period := time.Duration(100)
					appendEntriesArgs := make([]*AppendEntriesArgs, len(peers))
					appendEntriesReply := make([]*AppendEntriesReply, len(peers))

					DPrintf("[server: %v]Leader, send heartbeat, period: %v\n", rf.me, period*time.Millisecond)
					for server := range peers {
						if server != rf.me {
							appendEntriesArgs[server] = &AppendEntriesArgs{
								Term:         rf.currentTerm,
								LeaderID:     rf.me,
								PrevLogIndex: rf.nextIndex[server] - 1,
								PrevLogTerm:  rf.logs[rf.nextIndex[server]-1].LogTerm,
								Entries:      nil,
								LeaderCommit: rf.commitIndex}

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
											DPrintf("abc:%v, server: %v reply: %v\n", rf, server, reply)
											detectAppendEntriesArgs := &AppendEntriesArgs{
												Term:         rf.currentTerm,
												LeaderID:     rf.me,
												PrevLogIndex: rf.nextIndex[server] - 1,
												PrevLogTerm:  rf.logs[rf.nextIndex[server]-1].LogTerm,
												Entries:      nil,
												LeaderCommit: rf.commitIndex}
											rf.mu.Unlock()
											detectReply := new(AppendEntriesReply)
											ok1 := rf.sendAppendEntries(server, detectAppendEntriesArgs, detectReply)
											rf.mu.Lock()
											if !ok1 {
												DPrintf("[server: %v]not receive from %v\n", rf.me, server)
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
													DPrintf("[server: %v]Leader change to follower1: drain timer\n", rf.me)
													<-rf.t.C
												}
												timeout := time.Duration(500 + rand.Int31n(400))
												rf.t.Reset(timeout * time.Millisecond)

												rf.mu.Unlock()
												return
											}
											if detectReply.Success {
												firstTermIndex = detectAppendEntriesArgs.PrevLogIndex + 1
												break
											}
											rf.nextIndex[server] = detectReply.FirstTermIndex
										}
										DPrintf("[server: %v]Consistency check: server: %v, firstTermIndex: %v", rf.me, server, firstTermIndex)
										forceAppendEntriesArgs := &AppendEntriesArgs{
											Term:         rf.currentTerm,
											LeaderID:     rf.me,
											PrevLogIndex: firstTermIndex - 1,
											PrevLogTerm:  rf.logs[firstTermIndex-1].LogTerm,
											Entries:      rf.logs[firstTermIndex:],
											LeaderCommit: rf.commitIndex}

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
													DPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
													<-rf.t.C
												}
												timeout := time.Duration(500 + rand.Int31n(400))
												rf.t.Reset(timeout * time.Millisecond)

												rf.mu.Unlock()
												return
											} else {
												DPrintf("[server: %v]successfully append entries: %v\n", rf.me, forceReply)
												rf.nextIndex[server] = len(rf.logs)
												rf.matchIndex[server] = forceAppendEntriesArgs.PrevLogIndex + len(forceAppendEntriesArgs.Entries)
												rf.mu.Unlock()
												rf.cond.Broadcast()
												return
											}
										} else {
											DPrintf("[server: %v]no reponse from %v\n", rf.me, server)
											rf.nextIndex[server] = len(rf.logs)
										}
									} else {
										rf.state = FOLLOWER
										rf.currentTerm = reply.Term

										// reset timer
										if !rf.t.Stop() {
											DPrintf("[server: %v]Leader change to follower2: drain timer\n", rf.me)
											<-rf.t.C
										}
										timeout := time.Duration(500 + rand.Int31n(400))
										rf.t.Reset(timeout * time.Millisecond)
									}
								}
								rf.mu.Unlock()
							}(server, appendEntriesArgs[server], appendEntriesReply[server])
						}
					}

					rf.mu.Unlock()
					time.Sleep(period * time.Millisecond)

				}
			}
		}
	}(rf)

	go func(rf *Raft, applyCh chan ApplyMsg) {
		for {
			select {
			case <-rf.shutdown:
				DPrintf("[server: %v]Close logs handling goroutine\n", rf.me)
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
					DPrintf("[server: %v]matchIndex: %v, cntr: %v, rf.commitIndex: %v\n", rf.me, rf.matchIndex, matchIndexCntr, rf.commitIndex)
				}

				if rf.lastApplied < rf.commitIndex {
					DPrintf("[server: %v]lastApplied: %v, commitIndex: %v\n", rf.me, rf.lastApplied, rf.commitIndex)
					for rf.lastApplied < rf.commitIndex {
						rf.lastApplied++
						applyMsg := ApplyMsg{
							CommandValid: true,
							Command:      rf.logs[rf.lastApplied].LogCmd,
							CommandIndex: rf.lastApplied}
						DPrintf("[server: %v]send committed log to service: %v\n", rf.me, applyMsg)
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
