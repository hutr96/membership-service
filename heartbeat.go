package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func heartbeatTimeoutMonitor(hostname string, monitor chan bool) {
	// timeout monitor for hostname *
	fmt.Printf("// timeout monitor for hostname %s\n", hostname)
	for {
		select {
		case _ = <-monitor:
			continue
		case <-time.After(time.Duration(2*timeout) * time.Second):
			fmt.Printf("******Peer %s not reachable******\n", hostname)

			if _, ok := membershipList[hostname]; ok {
				membershipList[hostname] = false
			} else {
				fmt.Printf("No longer monitor Peer %s\n", hostname)
				return
			}
			if isLeader {
				go delPeerMsg(hostname)
			}
			if hostname == leader {
				fmt.Println("Leader failed")
				if isNextLeader() {
					fmt.Println("I am the next leader!")
					go leaderChange()
				}
			}
			// printMembershipList()

		}
	}
}

func heartbeatListen() {
	portT := strconv.Itoa(portUDP)
	udp, _ := net.ListenPacket("udp", ":"+portT)
	for {
		// fmt.Println("Heartbeat listening...")
		buf := make([]byte, 1024)
		n, _, err := udp.ReadFrom(buf)
		if err != nil {
			continue
		}
		go handleHeartbeat(buf[:n])
	}

}

func handleHeartbeat(buf []byte) {
	msg := strings.Split(string(buf), ":")
	fmt.Printf("HeartBeat received from: %s\n", msg)
	if isLeader && !membershipList[msg[1]] { // new peer d
		go addPeerMsg(msg[1])
	} else if _, ok := membershipList[msg[1]]; ok { // reset the time
		membershipList[msg[1]] = true
		timeoutMonitor[msg[1]] <- true
	}
}

func heartbeatSend() {
	for {
		for member := range membershipList {
			if member == myhostname {
				continue
			}
			addr := fmt.Sprintf("%s:%d", member, portUDP)
			go heartbeats(addr)
		}
		time.Sleep(time.Duration(timeout) * time.Second)
	}
}

func sendJoin() {
	// portT := strconv.Itoa(portUDP)

	for _, hostnameT := range hostnames {
		addr := fmt.Sprintf("%s:%d", hostnameT, portUDP)
		conn, err := net.Dial("udp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		conn.Write([]byte("Heartbeat:" + myhostname))
	}
	fmt.Printf("Join boardcast sent \n")
}

func heartbeats(addr string) {
	conn, _ := net.Dial("udp", addr)
	defer conn.Close()
	conn.Write([]byte("Heartbeat:" + myhostname))
	// fmt.Printf("Heartbeat sent to %s\n", addr)
}
