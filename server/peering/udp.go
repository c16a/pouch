package peering

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/store"
	"net"
	"os"
)

func InitPeer(nodeId string, peerAddr string, s *store.Node) {
	err := dialPeer(nodeId, peerAddr)
	if err != nil {
		fmt.Println("Failed to dial peer:", err)
	}

	// This blocks
	go startPeeringServer(s)
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

	joinRequest := &JoinRequest{NodeId: nodeId, Addr: raftAddr}

	joinRequestBytes, err := json.Marshal(joinRequest)
	if err != nil {
		return err
	}

	_, err = conn.Write(joinRequestBytes)
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

// JoinRequest is an incoming request from another node
//
// The underlying store will then add the remote node into its list.
type JoinRequest struct {
	NodeId string `json:"nodeId"` // The identifier of the node which is trying to connect to the current node
	Addr   string `json:"addr"`   // The address at which the remote node is reachable over the Raft network
}

type JoinResponse struct {
	OK  bool  `json:"ok"`
	Err error `json:"err"`
}

func handleUdpConnection(conn *net.UDPConn, s *store.Node) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var req JoinRequest
		err = json.Unmarshal(buf[:n], &req)
		if err != nil {
			continue
		}

		fmt.Printf("incoming UDP packet from %s, length %d\n", addr.String(), n)

		joinResponse := &JoinResponse{
			OK: false,
		}
		if err := s.Join(req.NodeId, req.Addr); err != nil {
			joinResponse.Err = err
		} else {
			joinResponse.OK = true
		}

		responseBytes, err := json.Marshal(joinResponse)
		if err != nil {
			continue
		}

		if _, err := conn.WriteToUDP(responseBytes, addr); err != nil {
			fmt.Printf("failed to send join response: %s\n", err)
		}
	}
}
