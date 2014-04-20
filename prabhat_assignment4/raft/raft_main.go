package raft

import (
	zmq "github.com/pebbe/zmq4"
	"math/rand"
	"time"
)

const (
	CONFIG_FILE = "peer.json"
	// No of Raft SERVER we can make it variable but for this we have ti read json file in raft layer which contain no of raft server
	MAX_SERVER = 5

	// Their are three type of phase
	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2

	// There are four type of message
	VOTE_R     = 0
	APPEND_REQ = 1
	VOTE_G     = 2
	APPEND_RES = 3

	// Three type of commands receive from clients
	SET    = "0"
	GET    = "1"
	DELETE = "2"

	// Reply to client success or failure
	ERROR   = "-1"
	NOERROR = "-2"
)

var (
	// This variable contain the phase of the server
	phase int

	// This variable keep track how many vote receive message receive with success
	count int

	// This is raft server object which contain all info of raft
	server Replicator

	// Send true to this channel when this raft server become LEADER
	lead = make(chan bool, 1)

	heartbeat_interm = make(chan int, 1)
	candidate_interm = make(chan int, 1)
	// When we want to close this raft server enable this channel
	ex_done = make(chan bool, 1)
	// this chennel for replicating the response of receive message to send_m() function of raft.go
	done = make(chan bool, 1)
	// This channel says some server need to send some AppendEntries message
	send [MAX_SERVER]chan bool
	// this channel for checking of reply of AppendEntries message
	success [MAX_SERVER]chan bool

	// This socket is for connecting the server from client
	responder *zmq.Socket
)

// This is main funtion which initialize different channel for commuication between receive_m and send_m function
func Start(myid int, port string) {
	rand.Seed(time.Now().UTC().UnixNano())

	// Initialize the raft object
	server = New(myid /* config file */, CONFIG_FILE)

	// Initially phase is Follower
	phase = FOLLOWER
	count = 0

	// This function is for communication of external client
	go raft_test(port)

	// This function receive message from lower cluster layer
	go receive_m()

	// This methord does all logical computation for state of server
	go send_m()

	// This method sense what shold be the commit index
	go commit()

	// This method update the lastApplied index also apply the command on state machine
	go stateMachine()

	for i := 1; i <= MAX_SERVER; i = i + 1 {
		if i != myid {

			// If this server is leader then this methord send AppendEntries message to other server
			go Append_Entries(i)
		}
	}
	// Signal from testing program to close this server
	// When we want to close the server test program send the true in this channel
	<-ex_done
	// Give some time so that it finish its all tasks
	time.Sleep(time.Second * 3)
}
