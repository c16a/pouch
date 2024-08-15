package store

import (
	"github.com/c16a/pouch/sdk/commands"
	"time"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

// Store is the interface Raft-backed key-value stores must implement.
type Store interface {
	// Get returns the value for the given key.
	Get(cmd *commands.GetCommand) (string, error)

	// Set sets the value for the given key, via distributed consensus.
	Set(cmd *commands.SetCommand) error

	// Delete removes the given key, via distributed consensus.
	Delete(cmd *commands.DelCommand) error

	// Join joins the node, identitifed by nodeID and reachable at addr, to the cluster.
	Join(nodeID string, addr string) error
}
