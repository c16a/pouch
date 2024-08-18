package commands

import (
	"errors"
	"fmt"
	"strings"
)

type JoinResponse struct {
	OK  bool  `json:"ok"`
	Err error `json:"err"`
}

func (r JoinResponse) String() string {
	if r.Err != nil {
		return fmt.Sprintf("%s %s", Err, r.Err.Error())
	}
	return fmt.Sprintf("OK %v", r.OK)
}

var (
	ErrorInvalidDataType = errors.New("invalid data type")
	ErrorNotFound        = errors.New("not found")
)

type ErrorResponse struct {
	Err error
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("%s %s", Err, e.Err.Error())
}

type CountResponse struct {
	Count int
}

func (c *CountResponse) String() string {
	return fmt.Sprintf("%s %d", Count, c.Count)
}

type BooleanResponse struct {
	Value bool
}

func (b *BooleanResponse) String() string {
	return fmt.Sprintf("%s %v", Boolean, b.Value)
}

type StringResponse struct {
	Value string
}

func (s *StringResponse) String() string {
	return fmt.Sprintf("%s %s", String, s.Value)
}

type ListResponse struct {
	Values []string
}

func (l *ListResponse) String() string {
	var result string
	for i, value := range l.Values {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("(%d): %s", i, value)
	}
	return result
}

type AuthChallengeRequestCommand struct {
	Challenge string
	LineMessage
}

func NewAuthChallengeRequestCommand(line LineMessage) (*AuthChallengeRequestCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &AuthChallengeRequestCommand{
		Challenge:   parts[1],
		LineMessage: line,
	}, nil
}

func (a *AuthChallengeRequestCommand) GetAction() MessageType {
	return AuthChallengeRequest
}

func (a *AuthChallengeRequestCommand) String() string {
	return fmt.Sprintf("%s %s", AuthChallengeRequest, a.Challenge)
}

// JoinCommand is an incoming request from another node
//
// The underlying store will then add the remote node into its list.
type JoinCommand struct {
	NodeId string `json:"nodeId"` // The identifier of the node which is trying to connect to the current node
	Addr   string `json:"addr"`   // The address at which the remote node is reachable over the Raft network
	line   string
}

func (c *JoinCommand) GetAction() MessageType {
	return Join
}

func (c *JoinCommand) String() string {
	return fmt.Sprintf("%s %s %s", string(Join), c.NodeId, c.Addr)
}

func NewJoinCommand(nodeId string, addr string) string {
	return fmt.Sprintf("%s %s %s", string(Join), nodeId, addr)
}
