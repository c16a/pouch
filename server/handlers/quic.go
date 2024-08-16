package handlers

import (
	"bufio"
	"context"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/store"
	"github.com/quic-go/quic-go"
	"io"
	"log"
	"os"
	"strings"
)

func StartQuicListener(node *store.Node) {
	quicAddr := os.Getenv(env.QuicAddr)
	if quicAddr == "" {
		log.Fatalf("Environment variable %s not set", env.QuicAddr)
	}

	listener, err := quic.ListenAddr(quicAddr, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			continue
		}
		go handleQuicConnection(conn, node)
	}
}

func handleQuicConnection(conn quic.Connection, node *store.Node) {
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return
	}

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)

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
