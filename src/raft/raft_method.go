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
	// NOTICE: Your code here (2A).

	// // TODO: 这里为什么要上锁呢
	// rf.mu.Lock()
	// defer rf.mu.Unlock()

	term = rf.currentTerm
	isleader = rf.state == LEADER

	return term, isleader
}

func (rf *Raft) isLeader() bool {
	rf.rwmu.RLock()
	defer rf.rwmu.RUnlock()
	return rf.state == LEADER
}

func (rf *Raft) getLastIndex() int {
	return rf.logs[len(rf.logs)-1].LogIndex
}

func (rf *Raft) getLastTerm() int {
	return rf.logs[len(rf.logs)-1].LogTerm
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
	debugPrintf("[server: %v]Encode: rf currentTerm: %v, votedFor: %v, log:%v\n", rf.me, rf.currentTerm, rf.votedFor, rf.logs)
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
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).
	// if command received from client:
	// append entry to local log, respond after entry applied to state machine
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	switch rf.state {
	case LEADER:
		index = len(rf.logs)
		term = rf.currentTerm
		isLeader = true

		logEntry := new(LogEntry)
		logEntry.LogTerm = rf.currentTerm
		logEntry.Command = command

		rf.logs = append(rf.logs, *logEntry)

		debugPrintf("[server: %v]appendEntriesArgs entry: %v\n", rf.me, *logEntry)

		appendEntriesArgs := make([]*AppendEntriesArgs, len(rf.peers))
		appendEntriesReply := make([]*AppendEntriesReply, len(rf.peers))

		for server := range rf.peers {
			if server != rf.me {
				appendEntriesArgs[server] = &AppendEntriesArgs{
					Term:         rf.currentTerm,
					LeaderID:     rf.me,
					PrevLogIndex: rf.nextIndex[server] - 1,
					PrevLogTerm:  rf.logs[rf.nextIndex[server]-1].LogTerm,
					Entries:      []LogEntry{*logEntry},
					LeaderCommit: rf.commitIndex}

				rf.nextIndex[server] += len(appendEntriesArgs[server].Entries)
				debugPrintf("leader:%v, nextIndex:%v\n", rf.me, rf.nextIndex)

				appendEntriesReply[server] = new(AppendEntriesReply)

				//go func(rf *Raft, server int, args *AppendEntriesArgs, reply *AppendEntriesReply) {
				//    for {
				//        DPrintf("[server: %v]qwe: %v\n", rf.me, args.Entries[0].Command)
				//        trialReply := new(AppendEntriesReply)
				//        ok := rf.sendAppendEntries(server, args, trialReply)
				//        rf.mu.Lock()
				//        if rf.state != "Leader" {
				//            rf.mu.Unlock()
				//            return
				//        }
				//        if args.Term != rf.currentTerm {
				//            rf.mu.Unlock()
				//            return
				//        }
				//        if ok && trialReply.Success {
				//            reply.Term    = trialReply.Term
				//            reply.Success = trialReply.Success
				//            rf.matchIndex[server] = appendEntriesArgs[server].PrevLogIndex + len(appendEntriesArgs[server].Entries)
				//            DPrintf("leader:%v, matchIndex:%v\n", rf.me, rf.matchIndex)
				//            rf.mu.Unlock()
				//            break
				//        }
				//        if ok && trialReply.Term > rf.currentTerm {
				//            rf.state = "Follower"
				//            rf.currentTerm = trialReply.Term
				//            rf.mu.Unlock()
				//            return
				//        }
				//        rf.mu.Unlock()
				//        time.Sleep(500 * time.Millisecond)
				//    }
				//    DPrintf("[server: %v]AppendEntries reply of %v from follower %v, reply:%v\n", rf.me, args, server, reply);
				//    rf.cond.Broadcast()
				//
				//}(rf, server, appendEntriesArgs[server], appendEntriesReply[server])
			}
		}

	default:
		isLeader = false
	}

	debugPrintf("[server: %v] return value: log index:%v, term:%v, isLeader:%v\n", rf.me, index, term, isLeader)
	return index, term, isLeader
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
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()
	close(rf.shutdown)
	rf.cond.Broadcast()
}

func (rf *Raft) hasShutdown() bool {
	select {
	case <-rf.shutdown:
		debugPrintf("[server: %v]Close logs handling goroutine\n", rf.me)
		return true
	default:
		return false
	}
}

func (rf *Raft) electionTimerReset() {
	timeout := time.Duration(500 + rand.Int31n(400))
	rf.electionTimer.Reset(timeout * time.Millisecond)
}
