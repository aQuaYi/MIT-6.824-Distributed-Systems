package raft

// RequestVoteArgs 获取投票参数
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// NOTICE: Your data here (2A, 2B).

	Term         int // candidate's term
	CandidateID  int // candidate requesting vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm  int // term of candidate's last log entry
}

// RequestVoteReply 投票回复
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// NOTICE: Your data here (2A).
	Term        int  // 投票人的 currentTerm
	VoteGranted bool // 返回 true，表示获得投票
}

// RequestVote 投票工作
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// NOTICE: Your code here (2A, 2B).
	reply.Term = rf.currentTerm
	reply.VoteGranted = false

	if args.Term < rf.currentTerm {
		return
	}

	if (rf.votedFor == -1 || rf.state == CANDIDATE) &&
		rf.lastApplied <= args.LastLogIndex && rf.currentTerm <= args.LastLogTerm {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateID
	}
}
