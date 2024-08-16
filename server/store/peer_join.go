package store

import (
	"encoding/json"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"net"
)

func handlePeerJoin(buf []byte, n int, s *RaftNode, conn *net.UDPConn, addr *net.UDPAddr) {
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
