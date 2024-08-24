package handlers

import (
	"bufio"
	"crypto/tls"
	"github.com/c16a/pouch/sdk/auth"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/store"
	"go.uber.org/zap"
	"io"
	"net"
	"strings"
)

func StartTcpListener(node *store.RaftNode) {
	logger := node.GetLogger()
	if node.Config.Tcp != nil && node.Config.Tcp.Enabled {
		startNetListener(node, "tcp", node.Config.Tcp.Addr)
	} else {
		logger.Warn("skipping TCP listener")
	}
}

func StartUnixListener(node *store.RaftNode) {
	logger := node.GetLogger()
	if node.Config.Unix != nil && node.Config.Unix.Enabled {
		startNetListener(node, "unix", node.Config.Unix.Path)
	} else {
		logger.Warn("skipping Unix listener")
	}
}

func startNetListener(node *store.RaftNode, protocol string, addr string) {
	logger := node.GetLogger()

	var listener net.Listener
	var err error
	tlsConfig, err := GetTlsConfig(node.Config)
	if err != nil {
		logger.Error("failed to load TLS config", zap.Error(err))
		return
	} else {
		if tlsConfig != nil {
			listener, err = tls.Listen(protocol, addr, tlsConfig)
		} else {
			listener, err = net.Listen(protocol, addr)
		}
	}
	if err != nil {
		logger.Error("failed to start net listener", zap.String("protocol", protocol), zap.Error(err))
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleNetConnection(conn, node)
	}
}

func handleNetConnection(conn net.Conn, node *store.RaftNode) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	authenticator := auth.NewChallengeAuthenticator(node)
	err := authenticator.Authenticate(reader, writer)
	if err != nil {
		response := (&commands.ErrorResponse{Err: err}).String()
		writer.WriteString(response + "\n")
		writer.Flush()
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		cmd, err := commands.ParseStringIntoCommand(line)
		if err != nil {
			continue
		}

		response := node.ApplyCmd(cmd)

		writer.WriteString(response + "\n")
		writer.Flush()
	}
}
