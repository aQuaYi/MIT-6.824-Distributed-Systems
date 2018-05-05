package raft

import (
	"fmt"
	"sort"
)

// ApplyMsg 是发送消息
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool        // TODO: 弄清楚这个干嘛的
	CommandIndex int         // Command zd Raft.logs 中的索引号
	Command      interface{} // Command 的具体内容
}

func (a ApplyMsg) String() string {
	return fmt.Sprintf("ApplyMsg{idx:%d,cmd:%v}", a.CommandIndex, a.Command)
}

// 每当 rf.logs 或 rf.commitIndex 有变化时，就收到通知
// 然后，检查发现有可以 commit 的 entry 的话
// 就通过 applyCh 发送 ApplyMsg 给 replication state machine 进行 commit
func (rf *Raft) checkApplyLoop(applyCh chan ApplyMsg) {
	rf.shutdownWG.Add(1)
	isChanged := false

	for {
		select {
		case <-rf.toCheckApplyChan:
			debugPrintf(" S#%d 在 checkApplyLoop 的 case <- rf.toCheckApplyChan，收到信号。将要检查是否有新的 entry 可以 commit", rf.me)
		case <-rf.shutdownChan:
			debugPrintf(" S#%d 在 checkApplyLoop 的 case <- rf.shutdownChan，收到信号。关闭 checkApplyLoop", rf.me)
			rf.shutdownWG.Done()
			return
		}

		rf.rwmu.Lock()

		// 如果 rf 是 LEADER
		// 先检查能否更新 rf.commitIndex
		if rf.state == LEADER {
			idx := maxMajorityIndex(rf.matchIndex)
			// paper 5.4.2, only log entries from the leader's current term are committed by counting replicas
			if idx > rf.commitIndex &&
				rf.logs[idx].LogTerm == rf.currentTerm {
				rf.commitIndex = idx
				debugPrintf("%s 发现了新的 maxMajorityIndex==%d, 已经更新 rf.commitIndex", rf, idx)
				isChanged = true
			}
		}

		if rf.lastApplied < rf.commitIndex {
			isChanged = true
			debugPrintf("%s 有新的 log 可以 commit ，因为 lastApplied(%d) < commitIndex(%d)", rf, rf.lastApplied, rf.commitIndex)
		}

		for rf.lastApplied < rf.commitIndex {
			rf.lastApplied++
			applyMsg := ApplyMsg{
				CommandValid: true, // TODO: ApplyMsg.CommandValid 属性是做什么用的
				Command:      rf.logs[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied}
			debugPrintf("%s apply %s", rf, applyMsg)
			applyCh <- applyMsg
		}

		if isChanged {
			rf.persist()
			isChanged = false
		}

		rf.rwmu.Unlock()
	}
}

// 返回 matchIndex 中超过半数的 Index
// 例如
// matchIndex == {8,7,6,5,4}
// 	     temp == {4,5,6,7,8}
// i = (5-1)/2 = 2
// 超过半数的 server 拥有 {4,5,6}
// 其中 temp[i] == 6 是最大值
func maxMajorityIndex(matchIndex []int) int {
	temp := make([]int, len(matchIndex))
	copy(temp, matchIndex)
	sort.Ints(temp)
	i := (len(matchIndex) - 1) / 2
	return temp[i]
}
