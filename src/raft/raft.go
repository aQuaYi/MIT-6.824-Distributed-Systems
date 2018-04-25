package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

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
const (
	// HBINTERVAL is haertbeat interval
	HBINTERVAL = 50 * time.Millisecond // 50ms
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
	log         []LogEntry // 此 server 中保存的 logs

	// Volatile state on all servers:
	commitIndex int // logs 中已经 commited 的 log 的最大索引号
	lastApplied int // logs 中最新元素的索引号

	// Volatile state on leaders:
	nextIndex  []int // 下一个要发送给 follower 的 log 的索引号
	matchIndex []int // leader 与 follower 共有的 log 的最大的索引号

	state     state
	voteCount int

	chanCommit    chan struct{}
	chanHeartbeat chan struct{}
	chanGrantVote chan struct{}
	chanLeader    chan struct{}
	chanApply     chan ApplyMsg
}

// GetState 可以获取 raft 对象的状态
// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// NOTICE: Your code here (2A).

	term = rf.currentTerm

	isleader = rf.state == LEADER

	return term, isleader
}

func (rf *Raft) getLastIndex() int {
	return rf.log[len(rf.log)-1].LogIndex
}

func (rf *Raft) getLastTerm() int {
	return rf.log[len(rf.log)-1].LogTerm
}

// IsLeader 反馈 rf 是否是 Leader
func (rf *Raft) IsLeader() bool {
	return rf.state == LEADER
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// NOTICE: Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// NOTICE: Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// Start 启动
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := rf.currentTerm
	isLeader := rf.state == LEADER
	if isLeader {
		index = rf.getLastIndex() + 1
		//fmt.Printf("raft:%d start\n",rf.me)
		rf.log = append(rf.log, LogEntry{LogTerm: term, LogCmd: command, LogIndex: index}) // append new entry from client
		rf.persist()
	}
	return index, term, isLeader
}

// Kill is
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// NOTICE: Your code here, if desired.
}

// rf 为自己拉票，以便赢得选举
func (rf *Raft) canvass() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateID:  rf.me,
		LastLogTerm:  rf.getLastTerm(),
		LastLogIndex: rf.getLastIndex(),
	}

	for i := range rf.peers {
		if i != rf.me && rf.state == CANDIDATE {
			go func(i int) {
				var reply RequestVoteReply
				rf.sendRequestVote(i, &args, &reply)

				// NOTICE: 后续如何处理
			}(i)
		}
	}

	return
}

// Make is
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// NOTICE: Your initialization code here (2A, 2B, 2C).

	rf.state = FOLLOWER
	rf.votedFor = -1
	rf.log = append(rf.log, LogEntry{})
	rf.currentTerm = 0

	rf.chanCommit = make(chan struct{}, 100)
	rf.chanHeartbeat = make(chan struct{}, 100)
	rf.chanGrantVote = make(chan struct{}, 100)
	rf.chanLeader = make(chan struct{}, 100)
	rf.chanApply = applyCh

	go func() {
		for {
			switch rf.state {
			case FOLLOWER:
				select {
				case <-rf.chanHeartbeat:
				case <-rf.chanGrantVote:
				case <-time.After(time.Millisecond * time.Duration(rand.Int63()%333+550)):
					rf.state = CANDIDATE
					rf.votedFor = -1
				}
			case LEADER:
				rf.boatcastAppendEntries()
				time.Sleep(HBINTERVAL)
			case CANDIDATE:
				rf.mu.Lock()
				rf.currentTerm++
				rf.votedFor = rf.me
				rf.voteCount = 1
				rf.persist()
				rf.mu.Unlock()
				go rf.canvass()
				select {
				case <-time.After(time.Millisecond * time.Duration(rand.Int63()%333+550)):
				case <-rf.chanHeartbeat:
					rf.state = FOLLOWER
				case <-rf.chanLeader:
					rf.mu.Lock()
					rf.state = LEADER
					rf.nextIndex = make([]int, len(rf.peers))
					rf.matchIndex = make([]int, len(rf.peers))
					for i := range rf.peers {
						rf.nextIndex[i] = rf.getLastIndex() + 1
					}
					rf.mu.Unlock()
				}
			}
		}
	}()

	go func() {
		for {

			select {
			case <-rf.chanCommit:
				rf.mu.Lock()
				commitIndex := rf.commitIndex
				baseIndex := rf.log[0].LogIndex
				for i := rf.lastApplied + 1; i <= commitIndex; i++ {
					msg := ApplyMsg{CommandIndex: i, Command: rf.log[i-baseIndex].LogCmd}
					applyCh <- msg
					//fmt.Printf("me:%d %v\n",rf.me,msg)
					rf.lastApplied = i
				}
				rf.mu.Unlock()
			}
		}
	}()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
