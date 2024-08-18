package auth

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/sdk/pouchkey"
	"github.com/c16a/pouch/server/store"
	"strings"
)

var (
	ErrUnknownClient       = errors.New("unknown client")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrNoRegisteredClients = errors.New("no registered clients")
)

type ChallengeAuthenticator struct {
	node *store.RaftNode
}

func NewChallengeAuthenticator(node *store.RaftNode) *ChallengeAuthenticator {
	return &ChallengeAuthenticator{node: node}
}

func (c *ChallengeAuthenticator) Authenticate(reader *bufio.Reader, writer *bufio.Writer) error {
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

	switch challengeResponseCommand.GetMessageType() {
	case commands.AuthChallengeResponse:
		cmd := challengeResponseCommand.(*commands.AuthChallengeResponseCommand)
		clientId := cmd.ClientId
		challengeSignature := cmd.ChallengeSignature

		clients := c.node.Config.Auth.Clients

		if clients == nil {
			return ErrNoRegisteredClients
		}

		client, ok := clients[clientId]
		if !ok {
			return ErrUnknownClient
		}

		verifyOk := pouchkey.VerifyWithPublicKey(client.HexPublicKey, challenge, challengeSignature)
		if !verifyOk {
			return ErrInvalidSignature
		}
		return nil
	default:
		return commands.ErrInvalidCommand
	}
}
