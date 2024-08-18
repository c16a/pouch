package auth

import (
	"bufio"
)

type Authenticator interface {
	Authenticate(reader *bufio.Reader, writer *bufio.Writer) error
}
