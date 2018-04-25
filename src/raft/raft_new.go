package raft

import (
	"labrpc"
	"sync"
	"time"
)

type state int

// 规定了 server 所需的 3 种状态
const (
	LEADER state = iota
	CANDIDATE
	FOLLOWER
)

func (s state) String() string {
	switch s {
	case LEADER:
		return "Leader"
	case CANDIDATE:
		return "Candidate"
	case FOLLOWER:
		return "Follower"
	default:
		panic("出现了第4种状态")
	}
}

const (
	// HBINTERVAL is haertbeat interval
	HBINTERVAL = 50 * time.Millisecond // 50ms

	// NULL 表示没有投票给任何人
	NULL = -1
)

// Raft implements a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// NOTICE: Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// Persistent state on call servers
	currentTerm int        // 此 server 当前所处的 term 编号
	votedFor    int        // 此 server 在此 term 中投票给了谁，是 peers 中的索引号
	logs        []LogEntry // 此 server 中保存的 logs

	// Volatile state on all servers:
	commitIndex int // logs 中已经 commited 的 log 的最大索引号
	lastApplied int // logs 中最新元素的索引号

	// Volatile state on leaders:
	nextIndex  []int // 下一个要发送给 follower 的 log 的索引号
	matchIndex []int // leader 与 follower 共有的 log 的最大的索引号

	state    state
	t        *time.Timer
	cond     *sync.Cond
	shutdown chan struct{}

	// state     state
	// voteCount int

	// 	chanCommit    chan struct{}
	// 	chanHeartbeat chan struct{}
	// 	chanGrantVote chan struct{}
	// 	chanLeader    chan struct{}
	// 	chanApply     chan ApplyMsg

}

func newRaft(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// currentTerm
	rf.currentTerm = 0
	// votedFor
	rf.votedFor = NULL
	//log
	rf.logs = make([]LogEntry, 1)

	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.state = FOLLOWER
	rf.cond = sync.NewCond(&rf.mu)
	rf.shutdown = make(chan struct{})

	return rf
}
