package main

import (
	raft "github.com/pebbe/raft"
	"os"
	"strconv"
)

func main() {
	id := os.Args[1]
	port := os.Args[2]
	myid, _ := strconv.Atoi(id)
	raft.Start(myid, port)
}
