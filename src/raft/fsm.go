package raft

import (
	"log"
)

type fsmState int
type fsmEvent string
type fsmHandler func(*Raft) fsmState

// 规定了 server 所需的 3 种状态
const (
	LEADER fsmState = iota
	CANDIDATE
	FOLLOWER
)

func (s fsmState) String() string {
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

func (rf *Raft) setState(newState fsmState) {
	rf.state = newState
}

func (rf *Raft) getState() fsmState {
	return rf.state
}

func (rf *Raft) addHandler(state fsmState, event fsmEvent, handler fsmHandler) {
	if _, ok := rf.handlers[state]; !ok {
		rf.handlers[state] = make(map[fsmEvent]fsmHandler, 10)
	}
	if _, ok := rf.handlers[state][event]; ok {
		debugPrintf("[警告] FSM 的状态 (%s) 的事件 (%s) 的处理方法，被覆盖", state, event)
	}
	rf.handlers[state][event] = handler
}

func (rf *Raft) call(event fsmEvent) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	oldState := rf.getState()

	if rf.handlers[oldState] == nil ||
		rf.handlers[oldState][event] == nil {
		log.Fatalf("[错误] FSM 的状态 (%s) 没有转换事件", rf.getState())
	}

	newState := rf.handlers[oldState][event](rf)
	rf.setState(newState)

	debugPrintf("[server: %d] [%s] 事件导致 server 从 [%s] 转变成 [%s]", rf.me, event, oldState, newState)
}
