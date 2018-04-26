package raft

import (
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

// AppendEntriesReply 是 flower 回复 leader 的内容
type AppendEntriesReply struct {
	Term           int  // 回复者的 term
	Success        bool // 返回 true，如果回复者满足 prevLogIndex 和 prevLogTerm
	FirstTermIndex int  // TODO: 这是干什么的？
}

// AppendEntries is
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// NOTICE: Your code here. (2A, 2B)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// defer rf.persist()

	DPrintf("[server: %v]Term:%v, server log:%v lastApplied %v, commitIndex: %v, received AppendEntries, %v, arg term: %v, arg log len:%v", rf.me, rf.currentTerm, rf.logs, rf.lastApplied, rf.commitIndex, args, args.Term, len(args.Entries))

	// 1. replay false at once if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	// TODO: WHY???
	rf.state = FOLLOWER

	if !rf.t.Stop() {
		DPrintf("[server: %v]AppendEntries: drain timer\n", rf.me)
		<-rf.t.C
	}
	timeout := time.Duration(500 + rand.Int31n(400))
	rf.t.Reset(timeout * time.Millisecond)

	// 2. false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(rf.logs) <= args.PrevLogIndex {
		DPrintf("[server: %v] log doesn't contain PrevLogIndex\n", rf.me)
		reply.Term = rf.currentTerm
		reply.FirstTermIndex = len(rf.logs)
		reply.Success = false
		rf.currentTerm = args.Term // TODO: why
		return
	}

	// 3. if an existing entry conflicts with a new one (same index but diff terms),
	//    delete the existing entry and all that follows it
	if rf.logs[args.PrevLogIndex].LogTerm != args.PrevLogTerm {
		DPrintf("[server: %v] log contains PrevLogIndex, but term doesn't match\n", rf.me)
		reply.FirstTermIndex = args.PrevLogIndex
		for rf.logs[reply.FirstTermIndex].LogTerm == rf.logs[reply.FirstTermIndex-1].LogTerm {
			if reply.FirstTermIndex > rf.commitIndex {
				reply.FirstTermIndex--
			} else {
				reply.FirstTermIndex = rf.commitIndex + 1
				break
			}
			DPrintf("[server: %v]FirstTermIndex: %v\n", rf.me, reply.FirstTermIndex)
		}
		rf.logs = rf.logs[:reply.FirstTermIndex]
		reply.Term = rf.currentTerm
		reply.Success = false
		rf.currentTerm = args.Term
		return
	}

	// 4. append any new entries not already in the log
	if len(args.Entries) == 0 {
		DPrintf("[server: %v]received heartbeat\n", rf.me)
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
