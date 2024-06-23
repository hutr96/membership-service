package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
)

// ReqMessage struct: a request id, current view id, operation type, and the id of the peer to be added
type ReqMessage struct {
	RequestID     int
	CurViewID     int
	OperationType string // 1: add, 2: remove
	PeerID        string
	MSG           string
}

// OKMessage struct: the request id and the current view id.
type OKMessage struct {
	RequestID int
	CurViewID int
}

// NewviewMessage : the new view id and new membership list to all the members including the new peer.
type NewviewMessage struct {
	RequestID int
	NewViewID int
	MemList   map[string]bool
}

// NewLeaderMessage : should contain request id, current view id, operation type; operation type is PENDING
type NewLeaderMessage struct {
	RequestID     int
	CurViewID     int
	OperationType string // ADD or DEL or NOTHING
	// MemList       []int
}

// Message : message
type Message struct {
	MSGTYPE      string // 1: Request msg, 2: OK message, 3: Newview Message, 4: NewLeader MSG
	Sender       string
	REQMSG       ReqMessage
	OKMSG        OKMessage
	NEWVIEWMSG   NewviewMessage
	NEWLEADERMSG NewLeaderMessage
}

/*
	TCP message listening
*/
func tcpListen() {

	tcp, _ := net.Listen("tcp", ":"+port)
	for {
		// accept connection
		fmt.Println("TCP listening...")
		conn, err := tcp.Accept()
		if err != nil {
			// error
			fmt.Println(err)
			os.Exit(1)
		}
		go handleMessage(conn) // keep listening
	}

}

/*
	TCP message sending
*/
func tcpSend(msg Message, hostname string) {
	var addr string
	addr = fmt.Sprintf("%s:%s", hostname, port)
	fmt.Printf("Sending %s msg to %s\n", msg.MSGTYPE, addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// error
		fmt.Println(err)
		os.Exit(1)
	}

	encoder := gob.NewEncoder(conn)
	encoder.Encode(msg)

	conn.Close()
}

/*
	return the number of alive peers
*/
func lenMembershipList() int {
	lenMembership := 0
	for _, alive := range membershipList {
		if alive {
			lenMembership++
		}
	}
	return lenMembership
}

/*
	print the membership list
*/
func printMembershipList() {
	fmt.Println("Membership List:")
	for mem, reachable := range membershipList {
		fmt.Printf("%s, %t\n", mem, reachable)
	}
	fmt.Println("----------------")
}

/*
	bool: if is next leader
*/
func isNextLeader() bool {
	return myhostname == lowestHostname()
}

/*
	return the lowest alive peer
*/
func lowestHostname() string {
	tempnumber := 100
	var thelowest string
	for mem, alive := range membershipList {
		if mynum, _ := strconv.Atoi(mem[4:]); alive && mynum < tempnumber {
			tempnumber = mynum
			thelowest = mem
		}
	}
	return thelowest
}
