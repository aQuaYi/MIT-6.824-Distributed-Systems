package raft

import "fmt"

type fsmState int
type fsmEvent string
type fsmHandler func(*Raft, interface{}) fsmState

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

func (rf *Raft) addHandler(state fsmState, event fsmEvent, handler fsmHandler) {
	if _, ok := rf.handlers[state]; !ok {
		rf.handlers[state] = make(map[fsmEvent]fsmHandler, 10)
	}
	if _, ok := rf.handlers[state][event]; ok {
		debugPrintf("[警告] FSM 的状态 (%s) 的事件 (%s) 的处理方法，被覆盖", state, event)
	}
	rf.handlers[state][event] = handler
}

func (rf *Raft) call(event fsmEvent, args interface{}) {
	rf.rwmu.Lock()
	defer rf.rwmu.Unlock()

	oldState := rf.state

	if rf.handlers[oldState] == nil ||
		rf.handlers[oldState][event] == nil {
		msg := fmt.Sprintf("# %s # 的状态 (%s) 没有事件 (%s) 的转换 handler", rf, oldState, event)
		panic(msg)
	}

	rf.state = rf.handlers[oldState][event](rf, args)

	debugPrintf("# %s # 事件 (%s) 发生，已从 [%s] → [%s]", rf, event, oldState, rf.state)
}
