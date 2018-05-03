package raft

import (
	"fmt"
)

// AppendEntriesArgs 是添加 log 的参数
type AppendEntriesArgs struct {
	Term         int // leader.currentTerm
	LeaderID     int // leader.me
	PrevLogIndex int // index of log entry immediately preceding new ones
	PrevLogTerm  int // term of prevLogIndex entry
	LeaderCommit int // leader.commitIndex

	Entries []LogEntry // 需要添加的 log 单元，为空时，表示此条消息是 heartBeat

}

func (a AppendEntriesArgs) String() string {
	return fmt.Sprintf("appendEntriesArgs{S#%d, term:%d, PrevLogIndex:%d, PrevLogTerm:%d, LeaderCommit:%d, entries:%v}",
		a.LeaderID, a.Term, a.PrevLogIndex, a.PrevLogTerm, a.LeaderCommit, a.Entries)
}

// AppendEntriesReply 是 flower 回复 leader 的内容
type AppendEntriesReply struct {
	Term      int  // 回复者的 term
	Success   bool // 返回 true，如果回复者满足 prevLogIndex 和 prevLogTerm
	NextIndex int  // 下一次发送的 AppendEntriesArgs.Entries[0] 在 Leader.logs 中的索引号
}

func (r AppendEntriesReply) String() string {
	return fmt.Sprintf("appendEntriesReply{term:%d, Success:%t, NextIndex:%d}",
		r.Term, r.Success, r.NextIndex)
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	return rf.peers[server].Call("Raft.AppendEntries", args, reply)
}

func newAppendEntriesArgs(leader *Raft, server int) *AppendEntriesArgs {
	prevLogIndex := leader.nextIndex[server] - 1
	return &AppendEntriesArgs{
		Term:         leader.currentTerm,
		LeaderID:     leader.me,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  leader.logs[prevLogIndex].LogTerm,
		Entries:      leader.logs[prevLogIndex+1:],
		LeaderCommit: leader.commitIndex,
	}
}

// AppendEntries is
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// NOTICE: Your code here. (2A, 2B)

	debugPrintf("%s receive %s", rf, args)

	reply.Term = rf.currentTerm
	reply.NextIndex = args.PrevLogIndex + 1

	// 1. Replay false at once if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}

	if args.Term > rf.currentTerm ||
		(rf.state == CANDIDATE && args.Term >= rf.currentTerm) {
		rf.call(discoverNewLeaderEvent,
			toFollowerArgs{
				term:     args.Term,
				votedFor: args.LeaderID,
			})
		// 这里不 return 是因为，接下来可以继续处理
	}

	// 运行到这里，可以认为接收到了合格的 rpc 信号，可以重置 election timer 了
	debugPrintf("%s 收到了 valid appendEntries RPC 信号，准备重置 election timer", rf)
	rf.resetElectionTimerChan <- struct{}{}

	// 把 lock 移动到 rf.call 的下面，避免死锁
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	// 2. Reply false at once if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(rf.logs) <= args.PrevLogIndex {
		debugPrintf("%s 含有的 logs 太短，不含有 PrevLogIndex == %d", rf, args.PrevLogIndex)
		reply.NextIndex = len(rf.logs)
		reply.Success = false
		return
	}

	// 3. if an existing entry conflicts with a new one (same index but diff terms),
	//    delete the existing entry and all that follows it
	if rf.logs[args.PrevLogIndex].LogTerm != args.PrevLogTerm {
		debugPrintf("%s 中 rf.logs[args.PrevLogIndex].LogTerm(%d)!=args.PrevLogTerm(%d) ", rf, rf.logs[args.PrevLogIndex].LogTerm, args.PrevLogTerm)
		reply.NextIndex = args.PrevLogIndex
		wrongTerm := rf.logs[args.PrevLogIndex].LogTerm

		for reply.NextIndex > rf.commitIndex+1 &&
			rf.logs[reply.NextIndex].LogTerm == wrongTerm {
			reply.NextIndex--
		}
		debugPrintf("%s reply.NextIndex == %d", rf, reply.NextIndex)

		// 删除失效的 logs
		rf.logs = rf.logs[:reply.NextIndex]
		reply.Success = false
		return
	}

	// 运行到这里，说明 rf.logs[args.PrevLogIndex].LogTerm == args.PrevLogTerm

	// 4. append any new entries not already in the log

	if len(args.Entries) == 0 {
		debugPrintf("%s  接收到 heartbeat", rf)
		reply.NextIndex = args.PrevLogIndex + 1
		reply.Success = false
		return
	}

	// 只保留合规的 logs
	rf.logs = rf.logs[:args.PrevLogIndex+1]
	debugPrintf("%s  len(rf.logs)== %d，准备 添加 entries{%v}, len(entries)==%d", rf, len(rf.logs), args.Entries, len(args.Entries))
	rf.logs = append(rf.logs, args.Entries...)
	debugPrintf("%s  len(rf.logs)== %d，已经 添加 entries{%v}, len(entries)==%d", rf, len(rf.logs), args.Entries, len(args.Entries))
	rf.persist()
	reply.NextIndex = len(rf.logs)
	reply.Success = true

	// 5. if leadercommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, len(rf.logs)-1)
	}

	rf.appendedNewEntriesChan <- struct{}{}
}
