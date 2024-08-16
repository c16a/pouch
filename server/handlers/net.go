package handlers

import (
	"bufio"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/store"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func StartTcpListener(node *store.Node) {
	tcpAddr := os.Getenv(env.TcpAddr)
	if tcpAddr == "" {
		log.Fatalf("Environment variable %s not set", env.TcpAddr)
	}

	startNetListener(node, "tcp", tcpAddr)
}

func StartUnixListener(node *store.Node) {
	unixAddr := os.Getenv(env.UnixAddr)
	if unixAddr == "" {
		log.Fatalf("Environment variable %s not set", env.UnixAddr)
	}

	startNetListener(node, "unix", unixAddr)
}

func startNetListener(node *store.Node, protocol string, addr string) {
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

func handleNetConnection(conn net.Conn, node *store.Node) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

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
