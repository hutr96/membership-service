package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var hostnames []string
var membershipList map[string]bool
var hostname string
var leader string
var isLeader bool
var myhostname string
var curReqID int
var curViewID int
var pendingReq map[int]ReqMessage
var pendingOK map[[2]int]map[string]bool
var pendingHostname map[[2]int]string
var testCase *int
var timeoutMonitor map[string](chan bool)
var reqMsg ReqMessage

const timeout int = 5

const port = "23456"

var portUDP = 23457

func main() {
	testCase = flag.Int("t", 0, "test case")
	hostfileAddr := flag.String("h", "hostfile", "hostfile")
	flag.Parse()

	// read host file
	f, _ := os.Open(*hostfileAddr)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		hostnames = append(hostnames, scanner.Text())
	}
	leader = hostnames[0] // set leader as the first host in the config file
	fmt.Println("leader:", hostnames[0])
	myhostname, _ = os.Hostname()
	fmt.Println("MyHostname:", myhostname)
	fmt.Println("Testcase:", *testCase)

	// initializations
	membershipList = make(map[string]bool)
	pendingReq = make(map[int]ReqMessage)
	pendingOK = make(map[[2]int]map[string]bool)
	pendingHostname = make(map[[2]int]string)
	timeoutMonitor = make(map[string](chan bool))

	go tcpListen()       // start listening msg
	go heartbeatListen() // start listening heartbeats

	if leader == myhostname {
		isLeader = true                   // if leader:
		membershipList[myhostname] = true // add leader to the mem list
	} else {
		isLeader = false
		sendJoin() // boardcast join message
	}
	go heartbeatSend() // start sending heartbeats

	// test cases handling:
	for {
		if *testCase == 2 && myhostname == hostnames[len(hostnames)-1] {
			time.Sleep(5 * time.Second)
			os.Exit(1)
		}
		if *testCase == 3 && myhostname != leader {
			ht, _ := strconv.Atoi(string(myhostname[len(myhostname)-1]))
			fmt.Println(ht)
			time.Sleep(time.Duration(5*ht) * time.Second)
			os.Exit(1)
		}
		if *testCase == 4 && myhostname == hostnames[len(hostnames)-1] {
			time.Sleep(6 * time.Second)
			os.Exit(1)
		}
	}
}

/*
	The function for handling received messages. It will check the message type and call the corresponding handling functions.
*/
func handleMessage(conn net.Conn) {

	dec := gob.NewDecoder(conn)
	msg := &Message{}
	dec.Decode(msg)
	fmt.Printf("Received %s msg: %v\n", msg.MSGTYPE, msg)
	conn.Close()
	if msg.MSGTYPE == "REQ" {
		go handleReqMsg(*msg)
	} else if msg.MSGTYPE == "OK" {
		go handleOKMsg(*msg)
	} else if msg.MSGTYPE == "NEWVIEW" {
		go handleNewViewMsg(*msg)
	} else if msg.MSGTYPE == "NEWLEADER" {
		go handleNewLeaderMsg(*msg)
	} else {
		fmt.Println("Unknown MSG type!!!!!!!!!!!")
	}
}

/*
	The delete peer function. Send del request to other peers
*/
func delPeerMsg(hostname string) {
	msg := Message{
		MSGTYPE: "REQ",
		Sender:  myhostname,
		REQMSG: ReqMessage{
			RequestID:     curReqID,
			CurViewID:     curViewID,
			OperationType: "DEL",
			PeerID:        hostname}}
	pendingHostname[[2]int{curReqID, curViewID}] = hostname
	pendingReq[curReqID] = msg.REQMSG
	for member, alive := range membershipList {
		// for testcase 4
		if *testCase == 4 && myhostname != "host2" && member == "host2" {
			continue
		}
		if alive {
			tcpSend(msg, member)
		}
	}
	if *testCase == 4 && myhostname == hostnames[0] { // for testcase 4
		log.Fatal("Test case 4: leader crash")
		os.Exit(1)
	}
}

/*
	The add peer function. Send add request to other peers
*/
func addPeerMsg(hostname string) {
	msg := Message{
		MSGTYPE: "REQ",
		Sender:  myhostname,
		REQMSG: ReqMessage{
			RequestID:     curReqID,
			CurViewID:     curViewID,
			OperationType: "ADD",
			PeerID:        hostname}}
	pendingHostname[[2]int{curReqID, curViewID}] = hostname
	pendingReq[curReqID] = msg.REQMSG
	for member, alive := range membershipList {

		if alive {
			tcpSend(msg, member)
		}
	}

}

/*
	The function for processing request message and send ok message back. Add the request message to the pendingReq list.
*/
func handleReqMsg(msg Message) {
	reqMsg := msg.REQMSG
	// curViewID = reqMsg.CurViewID
	if reqMsg.OperationType == "ADD" {
		pendingReq[reqMsg.RequestID] = reqMsg
		OKMsg := OKMessage{RequestID: reqMsg.RequestID, CurViewID: reqMsg.CurViewID}
		mgsToSend := Message{Sender: myhostname, MSGTYPE: "OK", OKMSG: OKMsg}
		go tcpSend(mgsToSend, msg.Sender)
	} else if reqMsg.OperationType == "DEL" {
		pendingReq[reqMsg.RequestID] = reqMsg
		OKMsg := OKMessage{RequestID: reqMsg.RequestID, CurViewID: reqMsg.CurViewID}
		mgsToSend := Message{Sender: myhostname, MSGTYPE: "OK", OKMSG: OKMsg}
		go tcpSend(mgsToSend, msg.Sender)
	}
}

/*
	The function for processing ok message. It will modify the membership list and send the new view message if receives ok messages from all alive peers
*/
func handleOKMsg(msg Message) {
	OKMsg := msg.OKMSG
	// initialization
	if len(pendingOK[[2]int{OKMsg.RequestID, OKMsg.CurViewID}]) == 0 {
		pendingOK[[2]int{OKMsg.RequestID, OKMsg.CurViewID}] = make(map[string]bool)
	}
	pendingOK[[2]int{OKMsg.RequestID, OKMsg.CurViewID}][msg.Sender] = true
	if len(pendingOK[[2]int{OKMsg.RequestID, OKMsg.CurViewID}]) >= (lenMembershipList()) { // received ok msg from all alive peers

		// changing the membership list
		if pendingReq[OKMsg.RequestID].OperationType == "ADD" {
			hostToAdd := pendingHostname[[2]int{OKMsg.RequestID, OKMsg.CurViewID}]
			delete(pendingHostname, [2]int{OKMsg.RequestID, OKMsg.CurViewID})
			membershipList[hostToAdd] = true
			fmt.Printf("Added new peer %s\n", hostToAdd)
		} else if pendingReq[OKMsg.RequestID].OperationType == "DEL" {
			hostToDel := pendingHostname[[2]int{OKMsg.RequestID, OKMsg.CurViewID}]
			delete(pendingHostname, [2]int{OKMsg.RequestID, OKMsg.CurViewID})
			delete(membershipList, hostToDel)
			fmt.Printf("Deleted peer %s\n", hostToDel)
		}

		// sending new view message to all peers
		curViewID++
		newMsg := Message{
			MSGTYPE: "NEWVIEW",
			Sender:  myhostname,
			NEWVIEWMSG: NewviewMessage{
				RequestID: OKMsg.RequestID,
				NewViewID: curViewID,
				MemList:   make(map[string]bool)}}

		newMsg.NEWVIEWMSG.MemList = membershipList
		for member := range membershipList {
			tcpSend(newMsg, member)
		}
	}

}

/*
	The function for processing newview message. Update view id and membershiplist. Print the new view. Clean some data structures.
*/
func handleNewViewMsg(msg Message) {
	newViewmsg := msg.NEWVIEWMSG
	curReqID = newViewmsg.RequestID + 1
	curViewID = newViewmsg.NewViewID
	membershipList = newViewmsg.MemList

	if pendingReq[newViewmsg.RequestID].OperationType == "ADD" || len(pendingReq) == 0 {
		for mem := range membershipList {
			if _, ok := timeoutMonitor[mem]; !ok && (mem != myhostname) {
				timeoutMonitor[mem] = make(chan bool)
				go heartbeatTimeoutMonitor(mem, timeoutMonitor[mem])
			}
		}
	} else if pendingReq[newViewmsg.RequestID].OperationType == "DEL" {
		delete(timeoutMonitor, pendingReq[newViewmsg.RequestID].PeerID)
	}
	delete(pendingReq, newViewmsg.RequestID)
	fmt.Printf("New view %d\n", curViewID)
	printMembershipList()
}

/*
	The function for processing new leader message. Respond pending request. Finish the received operation if is leader.
*/
func handleNewLeaderMsg(msg Message) {
	newleaderMsg := msg.NEWLEADERMSG
	if newleaderMsg.OperationType == "PENDING" {
		leader = msg.Sender
		// respond the pending request if has
		if len(pendingReq) != 0 {
			for _, reqMsg := range pendingReq {
				newmsg := Message{
					MSGTYPE: "NEWLEADER",
					Sender:  myhostname,
					NEWLEADERMSG: NewLeaderMessage{
						RequestID:     curReqID,
						CurViewID:     curViewID,
						OperationType: reqMsg.OperationType},
					REQMSG: reqMsg}
				tcpSend(newmsg, leader)
			}

		} else {
			newmsg := Message{
				MSGTYPE: "NEWLEADER",
				Sender:  myhostname,
				NEWLEADERMSG: NewLeaderMessage{
					RequestID:     curReqID,
					CurViewID:     curViewID,
					OperationType: "NOTHING"}}
			tcpSend(newmsg, leader)
		}

		// if has pending request, resend the request to all peers
	} else if newleaderMsg.OperationType == "ADD" {
		if val, ok := pendingHostname[[2]int{msg.REQMSG.RequestID, msg.REQMSG.CurViewID}]; val == msg.REQMSG.PeerID && ok {
			// if already processed the retransmission
		} else {
			msg := Message{
				MSGTYPE: "REQ",
				Sender:  myhostname,
				REQMSG:  msg.REQMSG}
			pendingHostname[[2]int{msg.REQMSG.RequestID, msg.REQMSG.CurViewID}] = msg.REQMSG.PeerID
			pendingReq[msg.REQMSG.RequestID] = msg.REQMSG
			for member := range membershipList {
				tcpSend(msg, member)
			}
		}
	} else if newleaderMsg.OperationType == "DEL" {
		if val, ok := pendingHostname[[2]int{msg.REQMSG.RequestID, msg.REQMSG.CurViewID}]; val == msg.REQMSG.PeerID && ok {
			// if already processed the retransmission
		} else {
			msg := Message{
				MSGTYPE: "REQ",
				Sender:  myhostname,
				REQMSG:  msg.REQMSG}
			pendingHostname[[2]int{msg.REQMSG.RequestID, msg.REQMSG.CurViewID}] = msg.REQMSG.PeerID
			pendingReq[msg.REQMSG.RequestID] = msg.REQMSG
			for member := range membershipList {
				tcpSend(msg, member)
			}
		}
	} else if newleaderMsg.OperationType == "NOTHING" {
		// do nothing?
	}

}

/*
	The function for starting leader changing. Send new leader message to all alive peers
*/
func leaderChange() {
	isLeader = true
	msg := Message{
		MSGTYPE: "NEWLEADER",
		Sender:  myhostname,
		NEWLEADERMSG: NewLeaderMessage{
			RequestID:     curReqID,
			CurViewID:     curViewID,
			OperationType: "PENDING"}}
	for mem, alive := range membershipList {
		if alive {
			tcpSend(msg, mem)
		}
	}

}
