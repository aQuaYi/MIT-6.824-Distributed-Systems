package raft

import (
	"fmt"
	"math/rand"
	"time"
)

// AppendEntriesArgs 是添加 log 的参数
type AppendEntriesArgs struct {
	Term         int // leader 的 term
	LeaderID     int // leader 的 ID
	PrevLogIndex int // index of log entry immediately preceding new ones
	PrevLogTerm  int // term of prevLogIndex entry

	Entries []LogEntry // 需要添加的 log 单元，为空时，表示此条消息是 heartBeat

	LeaderCommit int // leader 的 commitIndex
}

func (a AppendEntriesArgs) String() string {
	return fmt.Sprintf("server:%d, term:%d, PrevLogIndex:%d, PrevLogTerm:%d, LeaderCommit:%d, entries:%v",
		a.LeaderID, a.Term, a.PrevLogIndex, a.PrevLogTerm, a.LeaderCommit, a.Entries)
}

// AppendEntriesReply 是 flower 回复 leader 的内容
type AppendEntriesReply struct {
	Term           int  // 回复者的 term
	Success        bool // 返回 true，如果回复者满足 prevLogIndex 和 prevLogTerm
	FirstTermIndex int  // TODO: 这是干什么的？
}

// AppendEntries is
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// NOTICE: Your code here. (2A, 2B)
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()
	// defer rf.persist()

	debugPrintf("[server: %v]Term:%v, server log:%v lastApplied %v, commitIndex: %v, received AppendEntries, %v, arg term: %v, arg log len:%v", rf.me, rf.currentTerm, rf.logs, rf.lastApplied, rf.commitIndex, args, args.Term, len(args.Entries))

	// 1. replay false at once if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	// TODO: WHY???
	rf.state = FOLLOWER

	if !rf.electionTimer.Stop() {
		debugPrintf("[server: %v]AppendEntries: drain timer\n", rf.me)
		<-rf.electionTimer.C
	}
	timeout := time.Duration(500 + rand.Int31n(400))
	rf.electionTimer.Reset(timeout * time.Millisecond)

	// 2. false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(rf.logs) <= args.PrevLogIndex {
		debugPrintf("[server: %v] log doesn't contain PrevLogIndex\n", rf.me)
		reply.Term = rf.currentTerm
		reply.FirstTermIndex = len(rf.logs)
		reply.Success = false
		rf.currentTerm = args.Term // TODO: why
		return
	}

	// 3. if an existing entry conflicts with a new one (same index but diff terms),
	//    delete the existing entry and all that follows it
	if rf.logs[args.PrevLogIndex].LogTerm != args.PrevLogTerm {
		debugPrintf("[server: %v] log contains PrevLogIndex, but term doesn't match\n", rf.me)
		reply.FirstTermIndex = args.PrevLogIndex
		for rf.logs[reply.FirstTermIndex].LogTerm == rf.logs[reply.FirstTermIndex-1].LogTerm {
			if reply.FirstTermIndex > rf.commitIndex {
				reply.FirstTermIndex--
			} else {
				reply.FirstTermIndex = rf.commitIndex + 1
				break
			}
			debugPrintf("[server: %v]FirstTermIndex: %v\n", rf.me, reply.FirstTermIndex)
		}
		rf.logs = rf.logs[:reply.FirstTermIndex]
		reply.Term = rf.currentTerm
		reply.Success = false
		rf.currentTerm = args.Term
		return
	}

	// 4. append any new entries not already in the log
	if len(args.Entries) == 0 {
		debugPrintf("[server: %v]received heartbeat\n", rf.me)
	} else if len(rf.logs) == args.PrevLogIndex+1 {
		for i, entry := range args.Entries {
			rf.logs = append(rf.logs[:args.PrevLogIndex+i+1], entry)
		}
		// persist only when possible committed data
		// for leader, it's easy to determine
		// persist follower whenever update
		rf.persist()
	} else if len(rf.logs)-1 > args.PrevLogIndex &&
		len(rf.logs)-1 < args.PrevLogIndex+len(args.Entries) {
		for i := len(rf.logs); i <= args.PrevLogIndex+len(args.Entries); i++ {
			rf.logs = append(rf.logs[:i], args.Entries[i-args.PrevLogIndex-1])
		}
		// persist only when possible committed data
		// for leader, it's easy to determine
		// persist follower whenever update
		// NOTICE: 记得进行持久化工作
		rf.persist()
	}

	// 5. if leadercommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit < len(rf.logs)-1 {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = len(rf.logs) - 1
		}
		rf.cond.Broadcast()
	}

	rf.currentTerm = args.Term
	reply.Success = true
	reply.Term = rf.currentTerm
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	return rf.peers[server].Call("Raft.AppendEntries", args, reply)
}

func newAppendEntriesArgs(rf *Raft, server int) *AppendEntriesArgs {
	return &AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderID:     rf.me,
		PrevLogIndex: rf.nextIndex[server] - 1,
		PrevLogTerm:  rf.logs[rf.nextIndex[server]-1].LogTerm,
		Entries:      nil,
		LeaderCommit: rf.commitIndex,
	}
}

func newForceAppendEntriesArgs(rf *Raft, firstTermIndex int) *AppendEntriesArgs {
	return &AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderID:     rf.me,
		PrevLogIndex: firstTermIndex - 1,
		PrevLogTerm:  rf.logs[firstTermIndex-1].LogTerm,
		Entries:      rf.logs[firstTermIndex:],
		LeaderCommit: rf.commitIndex,
	}
}

// AppendEntries2 is
func (rf *Raft) AppendEntries2(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// NOTICE: Your code here. (2A, 2B)

	debugPrintf("[%s] receive appendEntriesArgs [%s]", rf, args)

	reply.Term = rf.currentTerm

	// 1. Replay false at once if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}

	if args.Term > rf.currentTerm {
		rf.call(discoverNewTermEvent, args.Term)
	}
	// 把 lock 移动到 rf.call 的下面，避免死锁

	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	// 2. Reply false at once if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(rf.logs) <= args.PrevLogIndex {
		debugPrintf("[%s] log doesn't contain PrevLogIndex\n", rf)
		reply.FirstTermIndex = len(rf.logs)
		reply.Success = false
		return
	}

	// 3. if an existing entry conflicts with a new one (same index but diff terms),
	//    delete the existing entry and all that follows it
	if rf.logs[args.PrevLogIndex].LogTerm != args.PrevLogTerm {
		debugPrintf("[server: %v] log contains PrevLogIndex, but term doesn't match\n", rf.me)
		reply.FirstTermIndex = args.PrevLogIndex
		// TODO: 简化这里的逻辑
		for rf.logs[reply.FirstTermIndex].LogTerm == rf.logs[reply.FirstTermIndex-1].LogTerm {
			if reply.FirstTermIndex > rf.commitIndex {
				reply.FirstTermIndex--
			} else {
				reply.FirstTermIndex = rf.commitIndex + 1
				break
			}
			debugPrintf("[server: %v]FirstTermIndex: %v\n", rf.me, reply.FirstTermIndex)
		}

		// 删除失效的 log
		rf.logs = rf.logs[:reply.FirstTermIndex]
		reply.Success = false
		return
	}

	// 运行到这里，说明 rf.logs[args.PrevLogIndex].LogTerm == args.PrevLogTerm

	// 4. append any new entries not already in the log

	if len(args.Entries) == 0 {
		debugPrintf("[%s] received heartbeat\n", rf)
	} else {
		rf.logs = rf.logs[:args.PrevLogIndex+1]
		rf.logs = append(rf.logs, args.Entries...)
		rf.persist()
	}

	// 5. if leadercommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, len(rf.logs)-1)
		// TODO: 发送通知到 检查 apply 的 groutine
	}

	reply.Success = true
}
