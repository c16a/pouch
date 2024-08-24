package store

import (
	"encoding/json"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"go.uber.org/zap"
	"net"
)

func handlePeerJoin(buf []byte, n int, s *RaftNode, conn *net.UDPConn, addr *net.UDPAddr) {
	joinResponse := &commands.JoinResponse{
		OK: false,
	}

	defer func() {
		responseBytes, err := json.Marshal(joinResponse)
		if err != nil {
			s.logger.Error("failed to marshal join response", zap.Error(err))
		}

		if _, err := conn.WriteToUDP(responseBytes, addr); err != nil {
			s.logger.Error("failed to write join response", zap.Error(err))
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
