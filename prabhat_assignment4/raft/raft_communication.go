package raft

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/syndtr/goleveldb/leveldb"
	"sort"
	s "strings"
	"time"
)

// This function is for check error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// This method sense what shold be the commit index
func commit() {
	for {
		// Function find the midean of match index
		time.Sleep(time.Millisecond * 50)
		temp := make([]int, len(server.matchIndex))
		for i := 0; i < len(temp); i = i + 1 {
			temp[i] = server.matchIndex[i]
		}
		sort.Ints(temp)
		server.commitIndex = temp[len(temp)/2]
	}
}

// This method update the lastApplied index also apply the command on state machine
func stateMachine() {
	// LevelDb for kv store
	db, err := leveldb.OpenFile("table1.db"+fmt.Sprintf("%d", server.Id), nil)
	Check(err)

	// Contain reply to client
	var reply string
	for {
		time.Sleep(time.Millisecond * 10)
		if server.lastApplied < server.commitIndex {
			server.lastApplied = server.lastApplied + 1
			command := s.Split(server.logEntry[server.lastApplied], " ")
			if command[2] == SET {
				value := ""
				for i := 4; i < len(command); i = i + 1 {
					if i < len(command)-1 {
						value = value + command[i] + " "
					} else {
						value = value + command[i]
					}
				}
				err = db.Put([]byte(command[3]), []byte(value), nil)
				reply = command[1] + " " + NOERROR
			} else if command[2] == GET {
				temp, err := db.Get([]byte(command[3]), nil)
				Check(err)
				reply = command[1] + " " + btos(temp)

			} else if command[2] == DELETE {
				err = db.Delete([]byte(command[3]), nil)
				reply = command[1] + " " + NOERROR
			} else {
				reply = command[1] + " " + ERROR
			}
			Check(err)

			// Only leader can be reply to client
			if phase == LEADER {
				responder.Send(reply, 0)
			}
		}
	}
}

// This methord is for communication of raft server to testing program using Response sockets and this server is close if
// receive msg is 0 if msg is 1 then return phase to testing program
func raft_test(port string) {
	var reply string
	count := 0
	responder, _ = zmq.NewSocket(zmq.DEALER)
	responder.Bind("tcp://*:" + port)
	for {
		msg, _ := responder.Recv(0)
		count = count + 1
		if msg == "0" {
			ex_done <- true
			break
		} else if msg == "1" {
			reply = "-1 " + fmt.Sprintf("%d", phase)
			responder.Send(reply, 0)
		} else if msg == "2" {
			reply = "-2 " + fmt.Sprintf("%d", server.leaderId)
			responder.Send(reply, 0)
		} else {
			if phase != LEADER {
				reply = "0 " + s.Split(msg, " ")[0] + " " + fmt.Sprintf("%d", server.leaderId)
				responder.Send(reply, 0)
			} else {

				server.lastLogIndex = server.lastLogIndex + 1
				term := server.currentTerm
				server.logEntry[server.lastLogIndex] = fmt.Sprintf("%d", term) + " " + msg
				server.nextIndex[server.Id-1] = server.nextIndex[server.Id-1] + 1
				server.matchIndex[server.Id-1] = server.matchIndex[server.Id-1] + 1
			}
		}
	}
}
