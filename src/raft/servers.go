package raft

type state int

// 规定了 server 所需的 3 种状态
const (
	LEADER state = iota
	CANDIDATE
	FOLLOWER
)
