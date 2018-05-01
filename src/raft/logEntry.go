package raft

// LogEntry is log entry
type LogEntry struct {
	LogIndex int         // 此 log 在 LEADER.logs 中的索引号
	LogTerm  int         // LEADER 在生成此 log 时的 LEADER.currentTerm
	Command  interface{} // 具体的命令内容
}

// TODO: 注释 truncateLog
func truncateLog(lastIncludedIndex int, lastIncludedTerm int, log []LogEntry) []LogEntry {
	var newLogEntries []LogEntry
	newLogEntries = append(newLogEntries, LogEntry{LogIndex: lastIncludedIndex, LogTerm: lastIncludedTerm})

	for index := len(log) - 1; index >= 0; index-- {
		if log[index].LogIndex == lastIncludedIndex && log[index].LogTerm == lastIncludedTerm {
			newLogEntries = append(newLogEntries, log[index+1:]...)
			break
		}
	}

	return newLogEntries
}
