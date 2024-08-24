package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/bbolt"
	"github.com/c16a/pouch/server/datatypes"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RaftNode is a simple key-value store, where all changes are made via Raft consensus.
type RaftNode struct {
	RaftDir  string
	RaftBind string

	mu sync.Mutex
	m  map[string]datatypes.Type // The key-value store for the system.

	raft *raft.Raft // The consensus mechanism

	logger *zap.Logger
	Config *NodeConfig
}

// NewRaftNode returns a new RaftNode.
func NewRaftNode(config *NodeConfig, logger *zap.Logger) *RaftNode {
	raftPath := config.Cluster.RaftDir
	if raftPath == "" {
		logger.Fatal("no raft dir specified")
	}

	raftAddr := config.Cluster.Addr
	if raftAddr == "" {
		logger.Fatal("no raft addr specified")
	}

	nodeId := config.Cluster.NodeID
	if nodeId == "" {
		id, err := uuid.NewV7()
		if err != nil {
			logger.Fatal("failed to generate node id", zap.Error(err))
		}
		config.Cluster.NodeID = id.String()
	}

	if err := os.MkdirAll(raftPath, 0700); err != nil {
		logger.Fatal("failed to create path for Raft storage", zap.Error(err))
	}

	return &RaftNode{
		RaftDir:  raftPath,
		RaftBind: raftAddr,
		m:        make(map[string]datatypes.Type),
		logger:   logger,
		Config:   config,
	}
}

func (node *RaftNode) GetLogger() *zap.Logger {
	return node.logger
}

// Start opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (node *RaftNode) Start() error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(node.Config.Cluster.NodeID)

	logWriter := &zapio.Writer{Log: node.logger, Level: zap.DebugLevel}

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", node.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(node.RaftBind, addr, 3, 10*time.Second, logWriter)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(node.RaftDir, retainSnapshotCount, logWriter)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	boltDB, err := bbolt.NewBoltStore(filepath.Join(node.RaftDir, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bbolt store: %s", err)
	}
	node.logger.Info("initialised logstore and stablestore")
	logStore := boltDB
	stableStore := boltDB

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, node, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	node.raft = ra

	peers := node.Config.Cluster.PeerAddrs
	if peers == nil || len(peers) == 0 {
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

	node.initPeer()

	return nil
}

func (node *RaftNode) ApplyCmd(cmd commands.Command) string {
	switch cmd.GetMessageType() {
	case commands.Get:
		return node.Get(cmd.(*commands.GetCommand))
	case commands.Set:
		return node.Set(cmd.(*commands.SetCommand))
	case commands.Del:
		return node.Delete(cmd.(*commands.DelCommand))
	case commands.LPush:
		return node.LPush(cmd.(*commands.LPushCommand))
	case commands.RPush:
		return node.RPush(cmd.(*commands.RPushCommand))
	case commands.LLen:
		return node.LLen(cmd.(*commands.LLenCommand))
	case commands.RPop:
		return node.RPop(cmd.(*commands.RPopCommand))
	case commands.LPop:
		return node.LPop(cmd.(*commands.LPopCommand))
	case commands.LRange:
		return node.LRange(cmd.(*commands.LRangeCommand))
	case commands.SAdd:
		return node.SAdd(cmd.(*commands.SAddCommand))
	case commands.SCard:
		return node.SCard(cmd.(*commands.SCardCommand))
	case commands.SIsMember:
		return node.SIsMember(cmd.(*commands.SIsMemberCommand))
	case commands.SMembers:
		return node.SMembers(cmd.(*commands.SMembersCommand))
	case commands.SInter:
		return node.SInter(cmd.(*commands.SInterCommand))
	case commands.SDiff:
		return node.SDiff(cmd.(*commands.SDiffCommand))
	case commands.SUnion:
		return node.SUnion(cmd.(*commands.SUnionCommand))
	case commands.PFAdd:
		return node.PFAdd(cmd.(*commands.PFAddCommand))
	case commands.PFCount:
		return node.PFCount(cmd.(*commands.PFCountCommand))
	default:
		return (&commands.ErrorResponse{Err: commands.ErrInvalidCommand}).String()
	}
}

// Get returns the value for the given key.
func (node *RaftNode) Get(cmd *commands.GetCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch {
		case val.GetName() == "string":
			strVal := val.(*datatypes.String)
			response := &commands.StringResponse{Value: strVal.GetValue()}
			return response.String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

// respondAfterRaftCommit is invoked for any command that needs consensus.
//
// For followers, it automatically relays the command to the current leader and returns their response.
//
// For leaders, it commits to its own Raft log and responds.
func (node *RaftNode) respondAfterRaftCommit(cmd commands.Command) string {
	if node.raft.State() != raft.Leader {
		return node.getResponseFromLeader(cmd)
	}

	b := []byte(cmd.String())

	f := node.raft.Apply(b, raftTimeout)
	if err := f.Error(); err != nil {
		response := &commands.ErrorResponse{Err: err}
		return response.String()
	}
	return f.Response().(string)
}

// Set sets the value for the given key.
func (node *RaftNode) Set(cmd *commands.SetCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

// Delete deletes the given key.
func (node *RaftNode) Delete(cmd *commands.DelCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) sendToLeaderViaUdp(cmd commands.Command, addr raft.ServerAddress, id raft.ServerID) (string, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", string(addr))
	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	_, err = conn.Write([]byte(cmd.String()))
	if err != nil {
		return "", err
	}

	var responseBytes = make([]byte, 1024)
	_, err = conn.Read(responseBytes)
	if err != nil {
		return "", err
	}
	return string(responseBytes), nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (node *RaftNode) Join(nodeID, addr string) error {
	node.logger.Info("incoming join request", zap.String("node_id", nodeID), zap.String("addr", addr))

	configFuture := node.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		node.logger.Error("failed to get raft configuration", zap.Error(err))
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				node.logger.Info("remote node already member of cluster, ignoring join request", zap.String("node_id", nodeID), zap.String("addr", addr))
				return nil
			}

			future := node.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing peer node", zap.String("node_id", nodeID), zap.String("addr", addr), zap.Error(err))
			}
		}
	}

	f := node.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	node.logger.Info("peer node joined successfully", zap.String("node_id", nodeID), zap.String("addr", addr))
	return nil
}

type fsm RaftNode

// Apply applies a Raft log entry to the key-value store.
//
// This command should only process the commands which mutate the key-value store
func (node *RaftNode) Apply(l *raft.Log) interface{} {
	cmd, err := commands.ParseStringIntoCommand(string(l.Data))
	if err != nil {
		return err
	}

	switch cmd.GetMessageType() {
	case commands.Set:
		return node.applySet(cmd.(*commands.SetCommand))
	case commands.Del:
		return node.applyDelete(cmd.(*commands.DelCommand))
	case commands.LPush:
		return node.applyLPush(cmd.(*commands.LPushCommand))
	case commands.RPush:
		return node.applyRPush(cmd.(*commands.RPushCommand))
	case commands.LPop:
		return node.applyLpop(cmd.(*commands.LPopCommand))
	case commands.RPop:
		return node.applyRpop(cmd.(*commands.RPopCommand))
	case commands.SAdd:
		return node.applySADD(cmd.(*commands.SAddCommand))
	case commands.PFAdd:
		return node.applyPFAdd(cmd.(*commands.PFAddCommand))
	default:
		node.logger.Error("unrecognised command", zap.String("type", string(cmd.GetMessageType())))
		return nil
	}
}

// Snapshot returns a snapshot of the key-value store.
func (node *RaftNode) Snapshot() (raft.FSMSnapshot, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	// Clone the map.
	o := make(map[string]datatypes.Type)
	for k, v := range node.m {
		o[k] = v
	}
	return &FsmSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (node *RaftNode) Restore(rc io.ReadCloser) error {
	o := make(map[string]datatypes.Type)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	node.m = o
	return nil
}

func (node *RaftNode) applySet(cmd *commands.SetCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.m[cmd.Key] = datatypes.NewString(cmd.Value)
	return (&commands.CountResponse{Count: 1}).String()
}

func (node *RaftNode) applyDelete(cmd *commands.DelCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	delete(node.m, cmd.Key)
	return (&commands.CountResponse{Count: 1}).String()
}

func (node *RaftNode) getResponseFromLeader(cmd commands.Command) string {
	leaderAddr, id := node.raft.LeaderWithID()
	response, err := node.sendToLeaderViaUdp(cmd, leaderAddr, id)
	if err != nil {
		errResponse := &commands.ErrorResponse{Err: err}
		return errResponse.String()
	}
	return response
}

func (node *RaftNode) initPeer() {
	peers := node.Config.Cluster.PeerAddrs

	if peers != nil && len(peers) > 0 {
		peerAddr := peers[0]
		if peerAddr != "" {
			err := node.dialPeer(peerAddr)
			if err != nil {
				node.logger.Error("failed to dial peer", zap.Error(err))
			}
		}
	}

	// This blocks
	go node.startPeeringServer()
}

func (node *RaftNode) dialPeer(peerAddr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	if node.Config.Cluster.Addr == "" {
		return errors.New("no raft address")
	}

	joinRequest, err := commands.NewJoinCommandWithValues(node.Config.Cluster.NodeID, node.Config.Cluster.Addr)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(joinRequest.String()))
	if err != nil {
		return err
	}

	return nil
}

// startPeeringServer will start a peering server and keep it open
func (node *RaftNode) startPeeringServer() error {
	peeringPort := node.Config.Cluster.Addr
	if peeringPort == "" {
		node.logger.Fatal("raft addr not specified")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", peeringPort)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	go handlePeerMessage(conn, node)
	return nil
}

func handlePeerMessage(conn *net.UDPConn, s *RaftNode) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		cmd, err := commands.ParseStringIntoCommand(string(buf[:n]))
		if err != nil {
			continue
		}

		switch cmd.GetMessageType() {
		case commands.Join:
			handlePeerJoin(buf, n, s, conn, addr)
		default:
			handlePeerLog(buf, n, s, conn, addr)
		}
	}
}

func handlePeerLog(buf []byte, n int, node *RaftNode, conn *net.UDPConn, addr *net.UDPAddr) {
	var strResponse string

	defer func() {
		if _, err := conn.WriteToUDP([]byte(strResponse), addr); err != nil {
			node.logger.Error("failed to write log response", zap.Error(err))
		}
	}()

	c, err := commands.ParseStringIntoCommand(string(buf[:n]))
	if err != nil {
		strResponse = (&commands.ErrorResponse{Err: err}).String()
		return
	}

	strResponse = node.ApplyCmd(c)
}
