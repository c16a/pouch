package peering

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/store"
	"net"
	"os"
)

func InitPeer(nodeId string, peerAddr string, node *store.Node) {
	err := dialPeer(nodeId, peerAddr)
	if err != nil {
		fmt.Println("Failed to dial peer:", err)
	}

	// This blocks
	go startPeeringServer(node)
}

func dialPeer(nodeId string, peerAddr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	raftAddr := os.Getenv(env.RaftAddr)
	if raftAddr == "" {
		return errors.New("no raft address")
	}

	joinRequest := commands.NewJoinCommand(nodeId, raftAddr)
	_, err = conn.Write([]byte(joinRequest))
	if err != nil {
		return err
	}

	return nil
}

// startPeeringServer will start a peering server and keep it open
func startPeeringServer(s *store.Node) error {
	peeringPort := os.Getenv(env.RaftAddr)
	if peeringPort == "" {
		return fmt.Errorf("environment variable %s not set", env.RaftAddr)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", peeringPort)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	go handleUdpConnection(conn, s)
	return nil
}

func handleUdpConnection(conn *net.UDPConn, s *store.Node) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		cmd, err := commands.ParseStringIntoCommand(string(buf[:n]))
		if err != nil {
			continue
		}

		switch cmd.GetAction() {
		case commands.Join:
			handleJoin(buf, n, s, conn, addr)
		default:
			handleLog(buf, n, s, conn, addr)
		}
	}
}

func handleLog(buf []byte, n int, node *store.Node, conn *net.UDPConn, addr *net.UDPAddr) {
	var strResponse string

	defer func() {
		if _, err := conn.WriteToUDP([]byte(strResponse), addr); err != nil {
			fmt.Println("Failed to write log response:", err)
		}
	}()

	c, err := commands.ParseStringIntoCommand(string(buf[:n]))
	if err != nil {
		strResponse = (&commands.ErrorResponse{Err: err}).String()
		return
	}

	switch c.GetAction() {
	case commands.Set:
		setCommand := c.(*commands.SetCommand)
		res := node.Set(setCommand)
		strResponse = res
	case commands.Del:
		delCommand := c.(*commands.DelCommand)
		res := node.Delete(delCommand)
		strResponse = res
	}
}

func handleJoin(buf []byte, n int, s *store.Node, conn *net.UDPConn, addr *net.UDPAddr) {
	joinResponse := &commands.JoinResponse{
		OK: false,
	}

	defer func() {
		responseBytes, err := json.Marshal(joinResponse)
		if err != nil {
			fmt.Println("Failed to marshal join response:", err)
		}

		if _, err := conn.WriteToUDP(responseBytes, addr); err != nil {
			fmt.Println("Failed to write join response:", err)
		}
	}()

	command, err := commands.ParseStringIntoCommand(string(buf[:n]))
	if err != nil {
		joinResponse.Err = err
		return
	}

	if joinCmd, ok := command.(*commands.JoinCommand); ok {
		if err := s.Join(joinCmd.NodeId, joinCmd.Addr); err != nil {
			joinResponse.Err = err
		} else {
			joinResponse.OK = true
		}
	} else {
		joinResponse.Err = fmt.Errorf("unknown command: %v", command)
	}
}
