package dialer

import (
	"bufio"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/sdk/pouchkey"
	"io"
	"net"
	"os"
	"strings"
)

func DialTcp(addr string, clientId string, encodedSeed string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		os.Exit(1)
	}

	defer conn.Close()

	handleTcpLoop(conn, clientId, encodedSeed)

}

func handleTcpLoop(conn net.Conn, clientId string, encodedSeed string) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	line = strings.TrimSpace(line)
	msg, err := commands.ParseStringIntoCommand(line)
	if err != nil {
		fmt.Println(err)
		return
	}

	if msg.GetMessageType() == commands.AuthChallengeRequest {
		authCmd := msg.(*commands.AuthChallengeRequestCommand)

		challenge := authCmd.Challenge
		signature, err := pouchkey.SignWithSeedAsHex(encodedSeed, challenge)
		if err != nil {
			fmt.Println(err)
			return
		}

		authChallengeResponse, err := commands.NewAuthChallengeResponseCommandWithValues(clientId, signature)
		if err != nil {
			fmt.Println(err)
			return
		}
		writer.WriteString(authChallengeResponse.String() + "\n")
		writer.Flush()
	} else {
		return
	}

	stdReader := bufio.NewReader(os.Stdin)

	for {
		line, err = stdReader.ReadString('\n')
		if err != nil {
			continue
		}

		line = strings.TrimSpace(line)
		writer.WriteString(line + "\n")
		writer.Flush()

		serverLine, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				break
			}
			continue
		}

		serverLine = strings.TrimSpace(serverLine)
		fmt.Println(serverLine)
	}
}
