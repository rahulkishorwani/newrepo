#Raft Implementation

`Raft` Raft is a consensus algorithm for managing replicated log. Here is implementation of this `Raft Algorithm`. In this algorithm one of the raft server work as a leader and here i choose that leader using this algorithm.
And the hole communication is monitor by this leader.

#Usage

(note: In the examples below, $ is the shell prompt, we use one argument raft id to run the server "----------------"
#### Code snippet in command line:
```
$ go run github.com/rahulkishorwani/newrepo/blob/master/prabhat_assignment4/main.go raft_server.id port_no
---------------------------------

```
For now we can communicate only one client and we can enter the port no of that client. In testing program i explain programetically how to connect a client to this raft cluster.

# Testing

Leader election related test cases i already made in my previous assignment here we only make some request to server and check whether its response is correct or not. 

```
go get github.com/rahulkishorwani/newrepo/tree/master/prabhat_assignment4/raft
go test github.com/rahulkishorwani/newrepo/tree/master/prabhat_assignment4
```

# The `cluster` package

`cluster` Cluster package use in this program is lower layer of raft which provide lower layer interface for seng and receive packets from ather raft server.

# The `raft` package

`raft` package contain 6 file all are for different purpose i'll give some little detail about these files.
```
`raft_main.go` This is the start or raft algorithm my main program feed the value of id and communication port here.
`raft_structure.go` This contain all the basic structure which this raft support.
`raft_interface.go` This interface communicate to lower cluster layer.
`raft_vote_request.go` This file is for send the Append Entry message.
`raft_communication.go` This is for communication with clients.
`raft.go` This contain main logic of raft algorithm.
```

### How it works

It just the golang implementation of raft algorithm's leader election part. Here we use a new function which initialize the raft object and make it able to communicate with lower cluster layer.
	The detail discription of other function are write with the function in program.
