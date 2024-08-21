package handlers

import (
	"bufio"
	"context"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/store"
	"github.com/quic-go/quic-go"
	"io"
	"log"
	"strings"
)

func StartQuicListener(node *store.RaftNode) {
	if node.Config.Quic == nil || !node.Config.Quic.Enabled {
		return
	}

	quicAddr := node.Config.Quic.Addr

	var listener *quic.Listener
	var err error
	tlsConfig, err := GetTlsConfig(node.Config)
	if err != nil {
		log.Fatal(err)
	} else {
		if tlsConfig != nil {
			listener, err = quic.ListenAddr(quicAddr, tlsConfig, nil)
		} else {
			log.Fatal("cannot start QUIC listener without TLSConfig")
		}
	}

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

func handleQuicConnection(conn quic.Connection, node *store.RaftNode) {
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
