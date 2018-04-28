package raft

import (
	"bytes"
	"encoding/gob"
)

// InstallSnapshotArgs is
type InstallSnapshotArgs struct {
	Term              int
	LeaderID          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

// InstallSnapshotReply is
type InstallSnapshotReply struct {
	Term int
}

// InstallSnapshot is
func (rf *Raft) InstallSnapshot(args InstallSnapshotArgs, reply *InstallSnapshotReply) {
	// Your code here.
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		return
	}
	// rf.chanHeartbeat <- struct{}{}
	rf.state = FOLLOWER

	// FIXME:
	// rf.currentTerm = rf.currentTerm

	// FIXME: don't have this method
	// rf.persister.SaveSnapshot(args.Data)

	rf.logs = truncateLog(args.LastIncludedIndex, args.LastIncludedTerm, rf.logs)

	// msg := ApplyMsg{UseSnapshot: true, Snapshot: args.Data}

	rf.lastApplied = args.LastIncludedIndex
	rf.commitIndex = args.LastIncludedIndex

	rf.persist()

	// rf.chanApply <- msg
}
func (rf *Raft) sendInstallSnapshot(server int, args InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	if ok {
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.state = FOLLOWER
			rf.votedFor = -1
			return ok
		}

		rf.nextIndex[server] = args.LastIncludedIndex + 1
		rf.matchIndex[server] = args.LastIncludedIndex
	}
	return ok
}

func (rf *Raft) readSnapshot(data []byte) {

	rf.readPersist(rf.persister.ReadRaftState())

	if len(data) == 0 {
		return
	}

	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)

	var LastIncludedIndex int
	var LastIncludedTerm int

	d.Decode(&LastIncludedIndex)
	d.Decode(&LastIncludedTerm)

	rf.commitIndex = LastIncludedIndex
	rf.lastApplied = LastIncludedIndex

	rf.logs = truncateLog(LastIncludedIndex, LastIncludedTerm, rf.logs)

	// msg := ApplyMsg{UseSnapshot: true, Snapshot: data}

	go func() {
		// rf.chanApply <- msg
	}()
}
