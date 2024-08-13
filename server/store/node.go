package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Node is a simple key-value store, where all changes are made via Raft consensus.
type Node struct {
	RaftDir  string
	RaftBind string

	mu sync.Mutex
	m  map[string]string // The key-value store for the system.

	raft *raft.Raft // The consensus mechanism

	logger *log.Logger
}

// New returns a new Node.
func New() *Node {
	return &Node{
		m:      make(map[string]string),
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (node *Node) Open(enableSingle bool, localID string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", node.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(node.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(node.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	boltDB, err := raftboltdb.New(raftboltdb.Options{
		Path: filepath.Join(node.RaftDir, "raft.db"),
	})
	if err != nil {
		return fmt.Errorf("new bbolt store: %s", err)
	}
	logStore := boltDB
	stableStore := boltDB

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, node, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	node.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

func (node *Node) ApplyCmd(cmd commands.Command) string {
	switch cmd.GetAction() {
	case commands.Get:
		return node.Get(cmd.(*commands.GetCommand))
	case commands.Set:
		return node.Set(cmd.(*commands.SetCommand))
	case commands.Del:
		return node.Delete(cmd.(*commands.DelCommand))
	default:
		return (&commands.ErrorResponse{Err: errors.New("unknown command")}).String()
	}
}

// Get returns the value for the given key.
func (node *Node) Get(cmd *commands.GetCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		response := &commands.StringResponse{Value: val}
		return response.String()
	} else {
		return (&commands.NilResponse{}).String()
	}
}

// Set sets the value for the given key.
func (node *Node) Set(cmd *commands.SetCommand) string {
	if node.raft.State() != raft.Leader {
		leaderAddr, id := node.raft.LeaderWithID()
		err := node.sendToLeaderViaUdp(cmd, leaderAddr, id)
		if err != nil {
			response := &commands.ErrorResponse{Err: err}
			return response.String()
		}
		response := &commands.CountResponse{Count: 1}
		return response.String()
	}

	b := []byte(cmd.String())

	f := node.raft.Apply(b, raftTimeout)
	if err := f.Error(); err != nil {
		response := &commands.ErrorResponse{Err: err}
		return response.String()
	}
	response := &commands.CountResponse{Count: 1}
	return response.String()
}

// Delete deletes the given key.
func (node *Node) Delete(cmd *commands.DelCommand) string {
	if node.raft.State() != raft.Leader {
		leaderAddr, id := node.raft.LeaderWithID()
		err := node.sendToLeaderViaUdp(cmd, leaderAddr, id)
		if err != nil {
			response := &commands.ErrorResponse{Err: err}
			return response.String()
		}
		response := &commands.CountResponse{Count: 1}
		return response.String()
	}

	b := []byte(cmd.String())

	f := node.raft.Apply(b, raftTimeout)
	if err := f.Error(); err != nil {
		response := &commands.ErrorResponse{Err: err}
		return response.String()
	}
	response := &commands.CountResponse{Count: 1}
	return response.String()
}

func (node *Node) sendToLeaderViaUdp(cmd commands.Command, addr raft.ServerAddress, id raft.ServerID) error {
	udpAddr, err := net.ResolveUDPAddr("udp", string(addr))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.Write([]byte(cmd.String()))
	if err != nil {
		return err
	}

	var responseBytes = make([]byte, 1024)
	_, err = conn.Read(responseBytes)
	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (node *Node) Join(nodeID, addr string) error {
	node.logger.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := node.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		node.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				node.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := node.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := node.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	node.logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

type fsm Node

// Apply applies a Raft log entry to the key-value store.
func (node *Node) Apply(l *raft.Log) interface{} {
	cmd, err := commands.ParseStringIntoCommand(string(l.Data))
	if err != nil {
		return err
	}

	switch cmd.GetAction() {
	case commands.Set:
		return node.applySet(cmd.(*commands.SetCommand))
	case commands.Del:
		return node.applyDelete(cmd.(*commands.DelCommand))
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", cmd.GetAction()))
	}
}

// Snapshot returns a snapshot of the key-value store.
func (node *Node) Snapshot() (raft.FSMSnapshot, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	// Clone the map.
	o := make(map[string]string)
	for k, v := range node.m {
		o[k] = v
	}
	return &FsmSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (node *Node) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	node.m = o
	return nil
}

func (node *Node) applySet(cmd *commands.SetCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.m[cmd.Key] = cmd.Value
	return nil
}

func (node *Node) applyDelete(cmd *commands.DelCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	delete(node.m, cmd.Key)
	return nil
}
