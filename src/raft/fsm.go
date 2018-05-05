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

	oldState := rf.state

	if rf.handlers[oldState] == nil ||
		rf.handlers[oldState][event] == nil {
		msg := fmt.Sprintf("%s  的状态 (%s) 没有事件 (%s) 的转换 handler", rf, oldState, event)
		panic(msg)
	}

	rf.state = rf.handlers[oldState][event](rf, args)

	debugPrintf("%s  事件(%s)，已从 (%s) → (%s)", rf, event, oldState, rf.state)

	rf.rwmu.Unlock()

}

// 全部的事件
var (
	// election timer 超时
	electionTimeOutEvent = fsmEvent("election timeout")
	// CANDIDATE 获取 majority 选票的时候，会触发此事件
	winElectionEvent = fsmEvent("win this term election")
	// server 在 RPC 通讯中，会首先检查对方的 term
	// 如果发现对方的 term 数比自己的高，会触发此事件
	discoverNewTermEvent = fsmEvent("discover new term")
	// server 在接收到 appendEntriesArgs 后，发现对方的 term
	discoverNewLeaderEvent = fsmEvent("disover a leader with higher term")
)

// 添加 rf 转换状态时的处理函数
func (rf *Raft) addHandlers() {

	// 添加 FOLLOWER 状态下的处理函数
	rf.addHandler(FOLLOWER, electionTimeOutEvent, fsmHandler(startNewElection))
	rf.addHandler(FOLLOWER, discoverNewTermEvent, fsmHandler(toFollower))
	rf.addHandler(FOLLOWER, discoverNewLeaderEvent, fsmHandler(toFollower))

	// 添加 CANDIDATE 状态下的处理函数
	rf.addHandler(CANDIDATE, electionTimeOutEvent, fsmHandler(startNewElection))
	rf.addHandler(CANDIDATE, winElectionEvent, fsmHandler(comeToPower))
	rf.addHandler(CANDIDATE, discoverNewTermEvent, fsmHandler(toFollower))
	rf.addHandler(CANDIDATE, discoverNewLeaderEvent, fsmHandler(toFollower))

	// 添加 LEADER 状态下的处理函数
	rf.addHandler(LEADER, discoverNewTermEvent, fsmHandler(toFollower))
	rf.addHandler(LEADER, discoverNewLeaderEvent, fsmHandler(toFollower))

}
