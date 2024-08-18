package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/sdk/pouchkey"
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

	err := handleAuthentication(reader, writer, node)
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

func handleAuthentication(reader *bufio.Reader, writer *bufio.Writer, node *store.RaftNode) error {
	// Start authentication
	challenge, err := pouchkey.NewChallenge(64)
	if err != nil {
		fmt.Println("Could not create authentication challenge")
		return err
	}

	challengeRequest := (&commands.AuthChallengeRequestCommand{Challenge: challenge}).String()

	writer.WriteString(challengeRequest + "\n")
	writer.Flush()

	challengeResponseLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	challengeResponseLine = strings.TrimSpace(challengeResponseLine)
	challengeResponseCommand, err := commands.ParseStringIntoCommand(challengeResponseLine)
	if err != nil {
		return err
	}

	switch challengeResponseCommand.GetAction() {
	case commands.AuthChallengeResponse:
		cmd := challengeResponseCommand.(*commands.AuthChallengeResponseCommand)
		clientId := cmd.ClientId
		challengeSignature := cmd.ChallengeSignature

		clients := node.Config.Auth.Clients

		if clients == nil {
			return errors.New("no clients found in configuration")
		}

		client, ok := clients[clientId]
		if !ok {
			return errors.New("unknown client")
		}

		verifyOk := pouchkey.VerifyWithPublicKey(client.HexPublicKey, challenge, challengeSignature)
		if !verifyOk {
			return errors.New("invalid signature")
		}
		return nil
	default:
		return errors.New("invalid challenge response action")
	}
}
