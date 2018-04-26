package raft

import "log"

// needDebug for Debugging
const needDebug = false

// debugPrintf 根据设置打印输出
func debugPrintf(format string, a ...interface{}) {
	if needDebug {
		log.Printf(format, a...)
	}
}
