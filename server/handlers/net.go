package handlers

import (
	"bufio"
	"github.com/c16a/pouch/sdk/auth"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/store"
	"io"
	"log"
	"net"
	"strings"
)

func StartTcpListener(node *store.RaftNode) {
	startNetListener(node, "tcp", node.Config.Tcp.Addr)
}

func StartUnixListener(node *store.RaftNode) {
	startNetListener(node, "unix", node.Config.Unix.Path)
}

func startNetListener(node *store.RaftNode, protocol string, addr string) {
	listener, err := net.Listen(protocol, addr)
	if err != nil {
		log.Fatal(err)
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
