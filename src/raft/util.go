package raft

import "log"

// needDebug for Debugging
const needDebug = true

// DPrintf 根据设置打印输出
func DPrintf(format string, a ...interface{}) (n int, err error) {
	if needDebug {
		log.Printf(format, a...)
	}
	return
}
