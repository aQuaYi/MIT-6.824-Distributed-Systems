package raft

import (
	"fmt"
	"labrpc"
	"sync"
	"time"
)

// Raft implements a single Raft peer.
type Raft struct {
	rwmu  sync.RWMutex        // Lock to protect shared access to this peer's state
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
	currentTerm int // 此 server 当前所处的 term 编号
	votedFor    int // 此 server 在此 term 中投票给了谁，是 peers 中的索引号
	// votedTerm   int        // 此 server 投票时所在的 term
	logs []LogEntry // 此 server 中保存的 logs

	// Volatile state on all servers:
	commitIndex int // logs 中已经 commited 的 log 的最大索引号
	lastApplied int // logs 中最新元素的索引号

	// Volatile state on leaders:
	nextIndex  []int // 下一个要发送给 follower 的 log 的索引号
	matchIndex []int // leader 与 follower 共有的 log 的最大的索引号

	// Raft 作为 FSM 管理自身状态所需的属性
	state    fsmState
	handlers map[fsmState]map[fsmEvent]fsmHandler

	//
	electionTimer *time.Timer // 超时，就由 FOLLOWER 变 CANDIDATE
	cond          *sync.Cond
	shutdown      chan struct{}
	// 当 rf 接收到合格的 rpc 信号时，会通过 receiveValidRPC 发送信号
	receiveValidRPC chan struct{}

	// candidate 或 leader 中途转变为 follower 的话，就关闭这个 channel 来发送信号
	// 因为，同一个 rf 不可能既是 candidate 又是 leader
	// 所以，用来通知的 channel 只要有一个就好了
	convertToFollowerChan chan struct{}

	//
}

func (rf *Raft) String() string {
	return fmt.Sprintf("server:%d, state:%s, term:%d, commitIndex:%d, lastApplied:%d, logs:%v",
		rf.me, rf.state, rf.currentTerm, rf.commitIndex, rf.lastApplied, rf.logs)
}

// func newRaft(peers []*labrpc.ClientEnd, me int,
// 	persister *Persister) *Raft {
// 	rf := &Raft{}
// 	rf.peers = peers
// 	rf.persister = persister
// 	rf.me = me
// 	// currentTerm
// 	rf.currentTerm = 0
// 	// votedFor
// 	rf.votedFor = NULL
// 	// logs
// 	rf.logs = make([]LogEntry, 1)
// 	rf.commitIndex = 0
// 	rf.lastApplied = 0
// 	rf.state = FOLLOWER
// 	rf.handlers = make(map[fsmState]map[fsmEvent]fsmHandler, 3)
// 	rf.cond = sync.NewCond(&rf.rwmu)
// 	rf.shutdown = make(chan struct{})
// 	// rf.heartbeat = make(chan struct{})
// 	timeout := time.Duration(150 + rand.Int31n(150))
// 	rf.electionTimer = time.NewTimer(timeout * time.Millisecond)
// 	return rf
// }

func newRaft(peers []*labrpc.ClientEnd, me int,
	persister *Persister) *Raft {
	rf := &Raft{
		peers:         peers,
		persister:     persister,
		me:            me,
		currentTerm:   0,
		votedFor:      NULL,
		logs:          make([]LogEntry, 1),
		commitIndex:   0,
		lastApplied:   0,
		state:         FOLLOWER,
		handlers:      make(map[fsmState]map[fsmEvent]fsmHandler, 3),
		shutdown:      make(chan struct{}),
		electionTimer: time.NewTimer(time.Second),
	}
	rf.cond = sync.NewCond(&rf.rwmu)
	rf.electionTimerReset()
	return rf
}

func newRaft2(peers []*labrpc.ClientEnd, me int,
	persister *Persister) *Raft {
	rf := &Raft{
		peers:         peers,
		persister:     persister,
		me:            me,
		currentTerm:   0,
		votedFor:      NULL,
		logs:          make([]LogEntry, 1),
		commitIndex:   0,
		lastApplied:   0,
		state:         FOLLOWER,
		handlers:      make(map[fsmState]map[fsmEvent]fsmHandler, 3),
		shutdown:      make(chan struct{}),
		electionTimer: time.NewTimer(time.Second),
	}
	rf.electionTimerReset()

	go electionTimeOutLoop(rf)

	return rf
}

//
func electionTimeOutLoop(rf *Raft) {
	for {
		if rf.hasShutdown() {
			return
		}

		select {
		case <-rf.electionTimer.C:
			rf.call(electionTimeOutEvent, nil)
		case <-rf.receiveValidRPC:
			rf.electionTimerReset()
		}
	}
}
