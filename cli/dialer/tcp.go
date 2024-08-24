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

	// Authentication handling as in your code
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
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

	// Create channels to handle communication between goroutines
	done := make(chan struct{})
	serverMessages := make(chan string)
	clientMessages := make(chan string)

	// Goroutine to handle reading from the server
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Println("Server error:", err)
				}
				close(done)
				return
			}
			line = strings.TrimSpace(line)
			serverMessages <- line
		}
	}()

	// Goroutine to handle reading from the standard input
	go func() {
		stdReader := bufio.NewReader(os.Stdin)
		for {
			line, err := stdReader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error:", err)
				continue
			}
			line = strings.TrimSpace(line)
			clientMessages <- line
		}
	}()

	// Main loop to handle messages from both server and stdin
	for {
		select {
		case msg := <-serverMessages:
			fmt.Printf(">%s\n", msg)
		case msg := <-clientMessages:
			writer.WriteString(msg + "\n")
			writer.Flush()
		case <-done:
			fmt.Println("Connection closed by server.")
			return
		}
	}
}
