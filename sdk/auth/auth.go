package auth

import "net"

type Authenticator interface {
	Authenticate(conn net.Conn) error
}
