package commands

import "fmt"

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
