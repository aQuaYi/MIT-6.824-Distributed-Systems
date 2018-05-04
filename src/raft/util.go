package raft

import "log"

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	debugPrintf("程序开始运行")
}

// needDebug for Debugging
const needDebug = true

// debugPrintf 根据设置打印输出
func debugPrintf(format string, a ...interface{}) {
	if needDebug {
		log.Printf(format, a...)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
