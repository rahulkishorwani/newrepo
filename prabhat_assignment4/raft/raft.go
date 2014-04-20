package raft

import (
	"fmt"
	cluster "github.com/pebbe/cluster"
	s "strings"
	"time"
)

// This methord take care of receive msg from other raft servers
func receive_m() {
	voteFor := 0
	var term int
	var m_type int

	for {
		envelope := <-server.Inbox() //If any massage in inbox
		m_type = envelope.M_type
		if m_type == VOTE_R {
			term = envelope.Msg.(cluster.RequestVote).Term
		} else if m_type == APPEND_REQ {
			term = envelope.Msg.(cluster.AppendEntries).Term
		} else if m_type == APPEND_RES {
			term = envelope.Msg.(cluster.AppendResponse).Term
		} else {
			term = envelope.Msg.(cluster.Reply).Term
		}
		if m_type == APPEND_REQ {
			msg := envelope.Msg.(cluster.AppendEntries)
			if term >= server.Get_Term() {
				heartbeat_interm <- msg.Term

				server.leaderId = msg.LeaderId
				if msg.Entries[msg.PrevLogIndex+1] == "" {
					if msg.PrevLogIndex > server.lastLogIndex {
						server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: APPEND_RES, Msg: cluster.AppendResponse{Term: server.currentTerm, Id: server.Id, LastLogIndex: server.lastLogIndex, LastLogTerm: server.lastLogTerm, Success: false}}
					} else {
						entryTerm := Index_to_Term(msg.PrevLogIndex)
						if entryTerm == msg.PrevLogTerm {
							index := msg.PrevLogIndex + 1
							for {
								if msg.Entries[index] != "" {
									if server.logEntry[index] != "" {
										// If leader is changed in between then this will tell the client to redirect to new leader

										reply := "0 " + s.Split(server.logEntry[index], " ")[1] + " " + fmt.Sprintf("%d", server.leaderId)
										responder.Send(reply, 0)
									}
									server.logEntry[index] = msg.Entries[index]
									index = index + 1
								} else {
									break
								}
							}
							server.lastLogIndex = index - 1
							server.lastLogTerm = Index_to_Term(server.lastLogIndex)
							server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: APPEND_RES, Msg: cluster.AppendResponse{Term: server.currentTerm, Id: server.Id, LastLogIndex: server.lastLogIndex, LastLogTerm: server.lastLogTerm, Success: true}}
							if server.lastLogIndex >= msg.LeaderCommit {
								server.commitIndex = msg.LeaderCommit
							}
						} else {
							server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: APPEND_RES, Msg: cluster.AppendResponse{Term: server.currentTerm, Id: server.Id, LastLogIndex: server.lastLogIndex, LastLogTerm: server.lastLogTerm, Success: false}}
						}
					}
				}
			} else {
				server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: APPEND_RES, Msg: cluster.AppendResponse{Term: server.currentTerm, Id: server.Id, LastLogIndex: server.lastLogIndex, LastLogTerm: server.lastLogTerm, Success: false}}
				continue
			}

		} else if m_type == APPEND_RES {
			timer1 := time.NewTimer(time.Nanosecond)
			select {
			//channel which says about re
			case <-success[envelope.Pid-1]:
			case <-timer1.C:
			}
			success[envelope.Pid-1] <- true

			msg := envelope.Msg.(cluster.AppendResponse)
			if msg.Success == false {
				if msg.Term > server.Get_Term() {
					candidate_interm <- term
				} else {
					lastterm := Index_to_Term(msg.LastLogIndex)
					if lastterm == msg.LastLogTerm {
						server.nextIndex[msg.Id-1] = msg.LastLogIndex + 1
						server.matchIndex[msg.Id-1] = msg.LastLogIndex
					} else {
						server.nextIndex[msg.Id-1] = server.nextIndex[msg.Id-1] - 1
					}
					continue
				}
			} else {
				server.nextIndex[msg.Id-1] = msg.LastLogIndex + 1
				server.matchIndex[msg.Id-1] = msg.LastLogIndex
				continue
			}
		} else if m_type == VOTE_R {
			msg := envelope.Msg.(cluster.RequestVote)
			if msg.LastLogIndex >= server.lastLogIndex {
				if term > server.Get_Term() {
					candidate_interm <- term
					voteFor = term
					server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: VOTE_G, Msg: cluster.Reply{Term: term, Success: true}}
				} else if term == server.Get_Term() {
					if phase == FOLLOWER {
						if voteFor < term {
							candidate_interm <- term
							voteFor = term
							server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: VOTE_G, Msg: cluster.Reply{Term: term, Success: true}}
						} else {
							continue
						}
					} else {
						continue
					}
				} else {
					server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: VOTE_G, Msg: cluster.Reply{Term: server.currentTerm, Success: false}}
					continue
				}
			} else {
				server.Outbox() <- &cluster.Envelope{Pid: envelope.Pid, M_type: VOTE_G, Msg: cluster.Reply{Term: server.currentTerm, Success: false}}
				continue
			}

			//vote grant
		} else if m_type == VOTE_G {
			msg := envelope.Msg.(cluster.Reply)
			if msg.Success == true {
				if server.Get_Term() == msg.Term {
					if phase == CANDIDATE {
						count = count + 1
						if count >= 3 {
							lead <- true
							time.Sleep(time.Millisecond)
						}
					}
				}
				continue
			} else {
				candidate_interm <- msg.Term
			}
		} else {
			continue
		}
		<-done
	}
}

// This function for send msg but most important this function implement leader election algo using receive_m function
func send_m() {

	for {
		if phase == LEADER {
			select {
			case term := <-heartbeat_interm:
				server.leader = false
				server.currentTerm = term
				phase = FOLLOWER
				done <- true
			case term := <-candidate_interm:
				server.leader = false
				server.currentTerm = term
				phase = FOLLOWER
				done <- true
			}
		} else if phase == FOLLOWER {
			// Election timeout timer
			timer1 := time.NewTimer(time.Millisecond * time.Duration(server.Timeout()))
			select {
			case term := <-heartbeat_interm:
				server.currentTerm = term
				done <- true
			case term := <-candidate_interm:
				server.currentTerm = term
				done <- true
			case <-timer1.C:
				phase = CANDIDATE
			}
		} else {
			timer1 := time.NewTimer(time.Millisecond * time.Duration(server.Timeout()))
			server.currentTerm = server.currentTerm + 1
			count = 1

			server.Outbox() <- &cluster.Envelope{Pid: -1, M_type: VOTE_R, Msg: cluster.RequestVote{Term: server.currentTerm, CandidateId: server.Id, LastLogIndex: server.lastLogIndex, LastLogTerm: server.lastLogTerm}}
			select {
			case term := <-heartbeat_interm:
				server.currentTerm = term
				phase = FOLLOWER
				done <- true
			case term := <-candidate_interm:
				server.leader = false
				server.currentTerm = term
				phase = FOLLOWER
				done <- true
			case <-lead:
				server.leader = true
				phase = LEADER
				server.leaderId = server.Id
				for i := 0; i < len(server.nextIndex); i++ {
					server.nextIndex[i] = server.lastLogIndex + 1
					server.matchIndex[i] = 0
				}
				done <- true
			case <-timer1.C:
				phase = FOLLOWER
				// Wait for some random time after candidate time out
			}
		}
	}
}
