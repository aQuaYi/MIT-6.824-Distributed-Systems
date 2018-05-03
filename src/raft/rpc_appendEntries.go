package raft

import (
	"fmt"
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
	Term      int  // 回复者的 term
	Success   bool // 返回 true，如果回复者满足 prevLogIndex 和 prevLogTerm
	NextIndex int  // 下一次发送的 AppendEntriesArgs.Entries[0] 在 Leader.logs 中的索引号
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

	debugPrintf("# %s # receive appendEntriesArgs [%s]", rf, args)

	reply.Term = rf.currentTerm

	// 1. Replay false at once if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}

	isArgsFromNewLeader := false

	rf.rwmu.RLock()
	if args.Term > rf.currentTerm ||
		(rf.state == CANDIDATE && args.Term >= rf.currentTerm) {
		isArgsFromNewLeader = true
	}
	rf.rwmu.RUnlock()

	if isArgsFromNewLeader {
		rf.call(discoverNewLeaderEvent,
			toFollowerArgs{
				term:     args.Term,
				votedFor: args.LeaderID,
			})
		// 这里不 return 是因为，接下来可以继续处理
	}

	// 运行到这里，可以认为接收到了合格的 rpc 信号，可以重置 election timer 了
	debugPrintf("# %s # 准备发送重置 election timer 信号", rf)
	rf.resetElectionTimerChan <- struct{}{}

	// 把 lock 移动到 rf.call 的下面，避免死锁
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	// 2. Reply false at once if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(rf.logs) <= args.PrevLogIndex {
		debugPrintf("# %s # log doesn't contain PrevLogIndex\n", rf)
		reply.NextIndex = len(rf.logs)
		reply.Success = false
		return
	}

	// 3. if an existing entry conflicts with a new one (same index but diff terms),
	//    delete the existing entry and all that follows it
	if rf.logs[args.PrevLogIndex].LogTerm != args.PrevLogTerm {
		debugPrintf("[server: %v] log contains PrevLogIndex, but term doesn't match\n", rf.me)
		reply.NextIndex = args.PrevLogIndex
		// TODO: 简化这里的逻辑
		for rf.logs[reply.NextIndex].LogTerm == rf.logs[reply.NextIndex-1].LogTerm {
			if reply.NextIndex > rf.commitIndex {
				reply.NextIndex--
			} else {
				reply.NextIndex = rf.commitIndex + 1
				break
			}
			debugPrintf("[server: %v]FirstTermIndex: %v\n", rf.me, reply.NextIndex)
		}

		// 删除失效的 log
		rf.logs = rf.logs[:reply.NextIndex]
		reply.Success = false
		return
	}

	// 运行到这里，说明 rf.logs[args.PrevLogIndex].LogTerm == args.PrevLogTerm

	// 4. append any new entries not already in the log

	if len(args.Entries) == 0 {
		debugPrintf("# %s # received heartbeat\n", rf)
	} else {
		rf.logs = rf.logs[:args.PrevLogIndex+1]
		rf.logs = append(rf.logs, args.Entries...)
		go rf.persist()
	}
	reply.NextIndex = len(rf.logs)

	// 5. if leadercommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, len(rf.logs)-1)
		// TODO: 发送通知到 检查 apply 的 groutine
	}

	reply.Success = true
}
