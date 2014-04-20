package raft

import (
	cluster "github.com/prabhat-bajpai/kvstore-public/tree/master/raft/cluster"
	"strconv"
	s "strings"
	"time"
)

func btos(k []byte) string {
	return string(k[:])
}

func stoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func iterator(key int) map[int]string {
	//----------------------------------------------------------------
	entry := make(map[int]string)
	count := 0
	for ok := key; ok <= server.lastLogIndex; ok++ {
		// Use key/value.
		entry[ok] = server.logEntry[ok]
		count = count + 1
		if count == 5 {
			break
		}
	}
	return entry

	//----------------------------------------------------------------
}

// For now i assume i have request of a particular pid and i have to send data only that id
// if it is leader then check the value of variable and send them append_Entries msg

func Index_to_Term(index int) int {

	return stoi(s.Split(server.logEntry[index], " ")[0])
}

// LEader send the appendEntries to other server
func Append_Entries(pid int) {
	// this function check whether send or not
	go Append_E2(pid)
	for {
		if phase == LEADER {
			if server.lastLogIndex >= server.nextIndex[pid-1] {
				timer1 := time.NewTimer(time.Nanosecond)
				select {
				//channel which says about re
				case <-send[pid-1]:
				case <-timer1.C:
				}
				send[pid-1] <- true
			}
			time.Sleep(time.Millisecond * 10)
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func Append_E2(pid int) {
	for {
		if phase == LEADER {
			timer1 := time.NewTimer(time.Millisecond * 200)
			prevLog := server.nextIndex[pid-1] - 1
			prevTerm := Index_to_Term(prevLog)
			entry := iterator(prevLog + 1)
			select {
			//channel which says about re
			case <-send[pid-1]:
				//Now i have to read all entry after prevLog
				//From Disk all entries after prevLog index
				msg := cluster.AppendEntries{Term: server.currentTerm, LeaderId: server.Id, PrevLogIndex: prevLog, PrevLogTerm: prevTerm, Entries: entry, LeaderCommit: server.commitIndex}
				//I put the msg in outbox and then wait for reply if no reply receive then resend and if fail receive then update variabes
				server.Outbox() <- &cluster.Envelope{Pid: pid, M_type: APPEND_REQ, Msg: msg}
				timer2 := time.NewTimer(time.Millisecond * 100)
				select {
				//channel which says about re
				case <-success[pid-1]:
				case <-timer2.C:
					timer3 := time.NewTimer(time.Nanosecond)
					select {
					//channel which says about re
					case <-send[pid-1]:
					case <-timer3.C:
					}
					send[pid-1] <- true
				}
			case <-timer1.C:
				msg := cluster.AppendEntries{Term: server.currentTerm, LeaderId: server.Id, PrevLogIndex: prevLog, PrevLogTerm: prevTerm, Entries: entry, LeaderCommit: server.commitIndex}

				server.Outbox() <- &cluster.Envelope{Pid: pid, M_type: APPEND_REQ, Msg: msg}
			}
			time.Sleep(time.Millisecond * 10)
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
}
