package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"os/exec"
	"strconv"
	s "strings"
	"testing"
	"time"
)

const (
	SET     = "0"
	GET     = "1"
	DELETE  = "2"
	ERROR   = "-1"
	NOERROR = "-2"
)

var (
	leaderId  int
	message   = make(map[int]string)
	requester = make([]*zmq.Socket, 5)
)

// This is simple send function here i communicate with raft using simple request-response socket and ask to raft their status
// like phase of raft and we can also close raft server using this function.
// Here msg :- 1 means msg for ask phase of server and msg:- 0 means close the raft server
func send(msg string, requester_C *zmq.Socket) string {

	// Decoding of phse receive by type 0 msg is 0 :- Follower, 1:- Candidate, 2:- Leader
	// By default phase :- "3" means no response from server and server is close
	requester_C.Send(msg, 0)
	phase := "3"
	if msg == "0" {
		return "0"
	}
	phase, _ = requester_C.Recv(0)
	type1 := s.Split(phase, " ")[0]
	value := s.Split(phase, " ")[1]
	if type1 == "-1" {
		if value == "0" {
			phase = "Follower"
		} else if value == "1" {
			phase = "Candidate"
		} else if value == "2" {
			phase = "Leader"
		}
	} else if type1 == "-2" {
		leaderId, _ = strconv.Atoi(value)
	} else if type1 == "0" {
		leaderId, _ = strconv.Atoi(s.Split(phase, " ")[2])
		msg_index, _ := strconv.Atoi(value)
		send(message[msg_index], requester[leaderId-1])
		return "0"
	} else {
	}
	fmt.Println(phase)
	return phase
}

// This function is used for start a Raft server from external program here we send the id of raft server
// and port for the response socket so that this program can communicate with server
func btos(k []byte) string {
	return string(k[:])
}
func execl(myid, port int) {
	lsCmd := exec.Command("go", "run", "main.go", fmt.Sprintf("%d", myid), fmt.Sprintf("%d", port))
	lsCmd.Stdout = os.Stdout
	lsCmd.Stdout = os.Stdout
	lsCmd.Start()
	lsCmd.Wait()
	//out, err := lsCmd.Output()
}

// This function check the phase of server at different test cases in each test case i change the configuration of server
// and run this method
// This method ask phase of Raft 100 times in 50 millisecond interval
func run(msg []string, up []bool, phase []string, t *testing.T) {
	// These variables for checking the status of whole custer

	for k := 1; k < 100; k++ {
		for i := 0; i < 5; i++ {
			if up[i] == true {
				phase[i] = send(msg[i], requester[i])
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
	for i := 0; i < 5; i++ {
		if up[i] == true {
			phase[i] = send("2", requester[i])
		}
	}
	time.Sleep(time.Millisecond * 50)
	db, _ := leveldb.OpenFile("table_test.db", nil)
	if leaderId != 0 {
		for k := 1; k <= 5; k++ {
			msg := fmt.Sprintf("%d", k) + " 0 " + fmt.Sprintf("%d", k) + " " + fmt.Sprintf("%d", k+1)
			_ = db.Put([]byte(fmt.Sprintf("%d", k)), []byte(fmt.Sprintf("%d", k+1)), nil)
			message[k] = NOERROR
			reply := send(msg, requester[leaderId-1])
			if message[k] != s.Split(reply, " ")[1] {
				t.Error("Wrong Output")
			}

			time.Sleep(time.Millisecond * 50)

		}
		for k := 1; k <= 5; k++ {
			msg := fmt.Sprintf("%d", k+5) + " 1 " + fmt.Sprintf("%d", k)
			temp, _ := db.Get([]byte(fmt.Sprintf("%d", k)), nil)
			message[k+5] = btos(temp)
			reply := send(msg, requester[leaderId-1])
			if message[k+5] != s.Split(reply, " ")[1] {
				t.Error("Wrong Output")
			}
			time.Sleep(time.Millisecond * 50)
		}
		for k := 1; k <= 5; k++ {
			msg := fmt.Sprintf("%d", k+10) + " 2 " + fmt.Sprintf("%d", k)
			_ = db.Delete([]byte(fmt.Sprintf("%d", k)), nil)
			message[k+10] = NOERROR
			reply := send(msg, requester[leaderId-1])
			if message[k+10] != s.Split(reply, " ")[1] {
				t.Error("Wrong Output")
			}
			time.Sleep(time.Millisecond * 50)
		}
	}
	time.Sleep(time.Second * 2)
}

func TestRaft(t *testing.T) {
	// This is the sockets for communicate with Raft server
	// This variable contain status of Raft server whether it 'up' or 'down'
	up := make([]bool, 5)
	// Msg for different servers
	msg := make([]string, 5) // 0 :- exit, 1:- what_phase
	// This is the phase of Different Raft server at that time
	phase := make([]string, 5)
	leaderId = 0
	port := 7232
	for i := 0; i < 5; i++ {
		port_i := port + i
		go execl(i+1, port_i)
		requester[i], _ = zmq.NewSocket(zmq.DEALER)
		requester[i].Connect("tcp://localhost:" + fmt.Sprintf("%d", port_i))
		up[i] = true
		msg[i] = "1"
	}

	// Test Case :- 1
	// 1st test case check one and only one leader at a time
	run(msg, up, phase, t)
	time.Sleep(time.Second * 20)
	// Test Case :- 2
	// close the leader
	for i := 0; i < 5; i++ {
		if leaderId == i+1 {
			send("0", requester[i])
			up[i] = false
			requester[i].Close()
		}
	}
	time.Sleep(time.Second * 2)
	run(msg, up, phase, t)
	time.Sleep(time.Second * 20)

	// Now down all the server and their respective process
	for i := 0; i < 5; i++ {
		if up[i] == true {
			send("0", requester[i])
			up[i] = false
			requester[i].Close()
		}
	}

	// Some Wait so is it can do their all task
	time.Sleep(time.Second * 2)
}
