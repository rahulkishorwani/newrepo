package raft

import cluster "github.com/pebbe/cluster"

// Identifies an entry in the log
type LogEntry struct {
	// An index into an abstract 2^64 size array
	Index int64
	//Term  int
	// The data that was supplied to raft's inbox
	Data interface{}
}

type Replicator struct {
	Id     int
	server cluster.Cluster // Cluster Object
	leader bool

	leaderId    int
	currentTerm int
	//votedFor    int	no need to make it global
	//log         []LogEntry

	commitIndex int
	lastApplied int
	//use only when this server become leader
	nextIndex  []int
	matchIndex []int // Be sure about the last log entry index replicated on server
	// currently i used this structure for log entry
	logEntry map[int]string
	//Index of last log entry in disk
	lastLogIndex int
	//Term of last log entry in disk
	lastLogTerm int
}
