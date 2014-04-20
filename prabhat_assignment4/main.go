package main

import (
	raft "github.com/rahulkishorwani/newrepo/tree/master/prabhat_assignment4/raft"
	"os"
	"strconv"
)

func main() {
	id := os.Args[1]
	port := os.Args[2]
	myid, _ := strconv.Atoi(id)
	raft.Start(myid, port)
}
