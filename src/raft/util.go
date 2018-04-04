package raft

import "log"

// Debug for Debugging
// 当 Debug > 0 时，会使得 DPrintf 输出信息
const Debug = 1

// DPrintf 根据设置打印输出
func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}
