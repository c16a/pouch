package env

const (
	NodeId   = "NODE_ID"
	TcpPort  = "TCP_PORT"
	RaftAddr = "RAFT_ADDR" // This is the advertised address for Raft peers
	PeerAddr = "PEER_ADDR" // This optionally lets the current node dial a peer during boot up
	RaftDir  = "RAFT_DIR"
	HttpAddr = "HTTP_ADDR"
)
