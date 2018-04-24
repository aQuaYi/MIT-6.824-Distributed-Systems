package raft

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
	Term      int  // 回复者的 term
	Success   bool // 返回 true，如果回复者满足 prevLogIndex 和 prevLogTerm
	NextIndex int
}

// AppendEntries is
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// Your code here.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	reply.Success = false
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.NextIndex = rf.getLastIndex() + 1
		//	fmt.Printf("%v currentTerm: %v rejected %v:%v\n",rf.me,rf.currentTerm,args.LeaderId,args.Term)
		return
	}
	rf.chanHeartbeat <- struct{}{}
	//fmt.Printf("%d respond for %v\n",rf.me,args.LeaderId)
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = FOLLOWER
		rf.votedFor = -1
	}
	reply.Term = args.Term

	if args.PrevLogIndex > rf.getLastIndex() {
		reply.NextIndex = rf.getLastIndex() + 1
		return
	}

	baseIndex := rf.log[0].LogIndex

	if args.PrevLogIndex > baseIndex {
		term := rf.log[args.PrevLogIndex-baseIndex].LogTerm
		if args.PrevLogTerm != term {
			for i := args.PrevLogIndex - 1; i >= baseIndex; i-- {
				if rf.log[i-baseIndex].LogTerm != term {
					reply.NextIndex = i + 1
					break
				}
			}
			return
		}
	}
	/*else {
		//fmt.Printf("????? len:%v\n",len(args.Entries))
		last := rf.getLastIndex()
		elen := len(args.Entries)
		for i := 0; i < elen ;i++ {
			if args.PrevLogIndex + i > last || rf.logs[args.PrevLogIndex + i].LogTerm != args.Entries[i].LogTerm {
				rf.log = rf.logs[: args.PrevLogIndex+1]
				rf.log = append(rf.log, args.Entries...)
				app = false
				fmt.Printf("?????\n")
				break
			}
		}
	}*/
	if args.PrevLogIndex < baseIndex {

	} else {
		rf.log = rf.log[:args.PrevLogIndex+1-baseIndex]
		rf.log = append(rf.log, args.Entries...)
		reply.Success = true
		reply.NextIndex = rf.getLastIndex() + 1
	}
	//println(rf.me,rf.getLastIndex(),reply.NextIndex,rf.log)
	if args.LeaderCommit > rf.commitIndex {
		last := rf.getLastIndex()
		if args.LeaderCommit > last {
			rf.commitIndex = last
		} else {
			rf.commitIndex = args.LeaderCommit
		}
		rf.chanCommit <- struct{}{}
	}
	return
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	return rf.peers[server].Call("Raft.AppendEntries", args, reply)
}
