package commands

import (
	"errors"
	"fmt"
	"strings"
)

type CommandKind string

// CommandAction is a request from client
type CommandAction string

const (
	Join   CommandAction = "JOIN"
	Get    CommandAction = "GET"
	Set    CommandAction = "SET"
	Del    CommandAction = "DEL"
	GetDel CommandAction = "GETDEL"

	LPush  CommandAction = "LPUSH"
	RPush  CommandAction = "RPUSH"
	LPop   CommandAction = "LPOP"
	RPop   CommandAction = "RPOP"
	LRange CommandAction = "LRANGE"
	LLen   CommandAction = "LLEN"
)

type Command interface {
	GetAction() CommandAction
	String() string
}

// JoinCommand is an incoming request from another node
//
// The underlying store will then add the remote node into its list.
type JoinCommand struct {
	NodeId string `json:"nodeId"` // The identifier of the node which is trying to connect to the current node
	Addr   string `json:"addr"`   // The address at which the remote node is reachable over the Raft network
	line   string
}

func (c *JoinCommand) GetAction() CommandAction {
	return Join
}

func (c *JoinCommand) String() string {
	return fmt.Sprintf("%s %s %s", string(Join), c.NodeId, c.Addr)
}

func NewJoinCommand(nodeId string, addr string) string {
	return fmt.Sprintf("%s %s %s", string(Join), nodeId, addr)
}

type GetCommand struct {
	Key  string
	line string
}

func (g *GetCommand) GetAction() CommandAction {
	return Get
}

func (g *GetCommand) String() string {
	return g.line
}

type SetCommand struct {
	Key   string
	Value string
	line  string
}

func (s *SetCommand) GetAction() CommandAction {
	return Set
}

func (s *SetCommand) String() string {
	return s.line
}

type DelCommand struct {
	Key  string
	line string
}

func (d *DelCommand) String() string {
	return d.line
}

func (d *DelCommand) GetAction() CommandAction {
	return Del
}

func ParseStringIntoCommand(s string) (Command, error) {
	parts := strings.Split(s, " ")

	if len(parts) == 0 {
		return nil, errors.New("invalid command")
	}

	action := parts[0]

	switch action {
	case string(Get):
		return &GetCommand{Key: parts[1], line: s}, nil
	case string(Set):
		return &SetCommand{Key: parts[1], Value: parts[2], line: s}, nil
	case string(Del):
		return &DelCommand{Key: parts[1], line: s}, nil
	case string(Join):
		return &JoinCommand{NodeId: parts[1], Addr: parts[2]}, nil
	default:
		return nil, errors.New("invalid command")
	}
}
