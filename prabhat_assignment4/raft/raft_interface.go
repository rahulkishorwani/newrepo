package raft

import 	cluster "github.com/prabhat-bajpai/kvstore-public/tree/master/raft/cluster"

// This is methods supported by Raft object
type Raft interface {
	Get_Term() int
	Get_Isleader() bool
	// Mailbox for state machine layer above to send commands of any
	// kind, and to have them replicated by raft.  If the server is not
	// the leader, the message will be silently dropped.
	Outbox() chan *cluster.Envelope

	//Mailbox for state machine layer above to receive commands. These
	//are guaranteed to have been replicated on a majority
	Inbox() chan *cluster.Envelope

	//Remove items from 0 .. index (inclusive), and reclaim disk
	//space. This is a hint, and there's no guarantee of immediacy since
	//there may be some servers that are lagging behind).
	//DiscardUpto(index int64)
}

// Methord Return the current term of the Raft
func (r Replicator) Get_Term() int {
	return r.currentTerm
}

// Methord Return true if is leader
func (r Replicator) Get_Isleader() bool {
	return r.leader
}

// Return Election time of this raft server
func (r Replicator) Timeout() int {
	return r.server.Timeout()
}

// Cluster layer Inbox
func (r Replicator) Inbox() chan *cluster.Envelope {
	return r.server.Inbox()
}

// Cluster layer Outnbox
func (r Replicator) Outbox() chan *cluster.Envelope {
	return r.server.Outbox()
}

// This Function initialize the parameters of Raft Object
func New(id int, f string) Replicator {
	server := cluster.New(id, f)
	nextindex := make([]int, MAX_SERVER)
	matchindex := make([]int, MAX_SERVER)
	for i := 0; i < MAX_SERVER; i++ {
		nextindex[i] = 1
		matchindex[i] = 0
	}
	first_log_index := 0
	last_Index := 0
	last_Term := 0
	count = 0
	for i := range send {
		send[i] = make(chan bool)
		success[i] = make(chan bool)

	}
	//handle if we restart server
	raft := Replicator{Id: id, server: server, currentTerm: last_Term, leader: false, commitIndex: first_log_index, lastApplied: first_log_index, nextIndex: nextindex, matchIndex: matchindex, logEntry: make(map[int]string), lastLogIndex: last_Index, lastLogTerm: last_Term}

	return raft
}
