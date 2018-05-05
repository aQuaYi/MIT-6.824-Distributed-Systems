package raft

import (
	"bytes"
	"io"
	"labgob"
	"math/rand"
	"time"
)

// GetState 可以获取 raft 对象的状态
// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (2A).

	// // TODO: 这里为什么要上锁呢
	// rf.mu.Lock()
	// defer rf.mu.Unlock()

	term = rf.currentTerm
	isleader = rf.state == LEADER

	return term, isleader
}

// TODO: 这个是有必要的吗
func (rf *Raft) getLastIndex() int {
	return rf.logs[len(rf.logs)-1].LogIndex
}

// TODO: 这个是有必要的吗
func (rf *Raft) getLastTerm() int {
	return rf.logs[len(rf.logs)-1].LogTerm
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
	//

	buffer := new(bytes.Buffer)
	e := labgob.NewEncoder(buffer)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	for i, b := range rf.logs {
		if i == 0 {
			continue
		}
		//if i > rf.commitIndex {
		//    break
		//}
		//DPrintf("[server: %v]Encode log: %v", rf.me, b)
		e.Encode(b.LogTerm)
		e.Encode(&b.Command)
	}
	data := buffer.Bytes()
	debugPrintf("%s Encode: rf currentTerm: %v, votedFor: %v, log:%v\n", rf, rf.currentTerm, rf.votedFor, rf.logs)
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
// func (*Deocder) Decode(e interface{}) error
//     Decode reads the next value from the input stream and stores it in
//     the data represented by the empty interface value. If e is nil, the
//     value will be discarded.
//     Otherwise, the value underlying e must be a pointer to the correct
//     type for the next data item received. If the input is at EOF,
//     Decode returns io.EOF and does not modify e
//
func (rf *Raft) readPersist(data []byte) {
	//DPrintf("[server: %v]read persist data: %v, len of data: %v\n", rf.me, data, len(data));
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
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

	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()
	buffer := bytes.NewBuffer(data)
	d := labgob.NewDecoder(buffer)
	var currentTerm int
	var votedFor int
	if d.Decode(&currentTerm) != nil ||
		d.Decode(&votedFor) != nil {
		debugPrintf("error in decode currentTerm and votedFor, err: %v\n", d.Decode(&currentTerm))
	} else {
		rf.currentTerm = currentTerm
		rf.votedFor = votedFor
	}
	for {
		var log LogEntry
		if err := d.Decode(&log.LogTerm); err != nil {
			if err == io.EOF {
				break
			} else {
				debugPrintf("error when decode log, err: %v\n", err)
			}
		}

		if err := d.Decode(&log.Command); err != nil {
			panic(err)
		}
		rf.logs = append(rf.logs, log)
	}
	//rf.commitIndex = len(rf.logs) - 1
	//rf.lastApplied = len(rf.logs) - 1
	debugPrintf("[server: %v]Decode: rf currentTerm: %v, votedFor: %v, log:%v, persist data: %v\n", rf.me, rf.currentTerm, rf.votedFor, rf.logs, data)

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
//
func (rf *Raft) Start(command interface{}) (index, term int, isLeader bool) {

	index = -1
	term = -1
	isLeader = false

	// Your code here (2B).
	// if command received from client:
	// append entry to local log, respond after entry applied to state machine
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	if rf.state != LEADER {
		return
	}

	// 修改结果值
	index = len(rf.logs)
	term = rf.currentTerm
	isLeader = true

	// 生成新的 entry
	entry := &LogEntry{
		LogIndex: index,
		LogTerm:  term,
		Command:  command,
	}

	// 修改 rf 的属性
	rf.logs = append(rf.logs, *entry)
	rf.nextIndex[rf.me] = len(rf.logs)
	rf.matchIndex[rf.me] = len(rf.logs) - 1

	debugPrintf("%s 添加了新的 entry:%v", rf, *entry)

	return
}

// Kill is
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	debugPrintf("S#%d Killing", rf.me)
	rf.toCheckApplyChan <- struct{}{}
	close(rf.shutdownChan)
	rf.shutdownWG.Wait()
}

func (rf *Raft) electionTimerReset() {
	timeout := time.Duration(150+rand.Int63n(151)) * time.Millisecond
	rf.electionTimer.Reset(timeout)
	debugPrintf("%s  election timer 已经重置，到期时间为 %s", rf, timeout)
}
