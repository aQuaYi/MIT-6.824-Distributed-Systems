package raft

import (
	"labrpc"
	"math/rand"
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
		panic("出现了第4种 server state")
	}
}

// Raft implements a single Raft peer.
type Raft struct {
	mu    sync.Mutex          // Lock to protect shared access to this peer's state
	peers []*labrpc.ClientEnd // RPC end points of all peers

	// Persistent state on all servers
	// (Updated on stable storage before responding to RPCs)
	// This implementation doesn't use disk; ti will save and restore
	// persistent state from a Persister object
	// Raft should initialize its state from Persister,
	// and should use it to save its persistent state each tiem the state changes
	// Use ReadRaftState() and SaveRaftState
	persister *Persister // Object to hold this peer's persisted state
	me        int        // this peer's index into peers[]

	// NOTICE: Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// from Figure 2

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

	// extra
	state    state
	t        *time.Timer
	cond     *sync.Cond
	shutdown chan struct{}
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

	timeout := time.Duration(300 + rand.Int31n(400))
	rf.t = time.NewTimer(timeout * time.Millisecond)
	return rf
}
