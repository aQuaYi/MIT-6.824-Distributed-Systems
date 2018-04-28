package raft

import (
	"math/rand"
	"time"
)

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

	// TODO: 注释这里的每一句话

	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	debugPrintf("[Enter RequestVote][server: %v]term :%v voted for:%v, log len: %v, logs: %v, commitIndex: %v, received RequestVote: %v\n", rf.me, rf.currentTerm, rf.votedFor, len(rf.logs), rf.logs, rf.commitIndex, args)

	// 1. false if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		// TODO: 此处直接 return 可否
	} else if args.Term > rf.currentTerm {
		rf.votedFor = NULL
		rf.state = FOLLOWER
		// TODO: 此处直接 return 可否
	}
	// TODO: Term 相等的情况是什么

	// 2. votedFor is null or candidateId and
	//    candidate's log is at least as up-to-date as receiver's log, then grant vote
	//    If the logs have last entries with different terms, then the log with the later term is more up-to-date
	//    If the logs end with the same term, then whichever log is longer is more up-to-date
	//
	if (rf.votedFor == NULL || rf.votedFor == args.CandidateID) &&
		((args.LastLogTerm > rf.logs[len(rf.logs)-1].LogTerm) ||
			((args.LastLogTerm == rf.logs[len(rf.logs)-1].LogTerm) && args.LastLogIndex >= len(rf.logs)-1)) {
		debugPrintf("[RequestVote][server: %v]term :%v voted for:%v, logs: %v, commitIndex: %v, received RequestVote: %v\n", rf.me, rf.currentTerm, rf.votedFor, rf.logs, rf.commitIndex, args)
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		rf.votedFor = args.CandidateID
		if !rf.electionTimer.Stop() {
			debugPrintf("[server %d] RequestVote: drain timer\n", rf.me)
			// TODO: 这是通知到什么地方了
			<-rf.electionTimer.C
		}
		timeout := time.Duration(500 + rand.Int31n(400))
		rf.electionTimer.Reset(timeout * time.Millisecond)
	} else {
		reply.VoteGranted = false
	}

}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) newRequestVoteArgs() *RequestVoteArgs {
	args := &RequestVoteArgs{
		Term:         rf.currentTerm + 1,
		CandidateID:  rf.me,
		LastLogIndex: len(rf.logs) - 1,
		LastLogTerm:  rf.logs[len(rf.logs)-1].LogTerm,
	}
	debugPrintf("[server: %v] Candidate,  send RequestVote: %v\n", rf.me, args)
	return args
}
