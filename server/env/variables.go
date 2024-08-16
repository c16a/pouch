package env

const (
	NodeId   = "NODE_ID"
	TcpAddr  = "TCP_ADDR"
	UnixAddr = "UNIX_ADDR"
	WsAddr   = "WS_ADDR"
	QuicAddr = "QUIC_ADDR"
	RaftAddr = "RAFT_ADDR" // This is the advertised address for Raft peers
	PeerAddr = "PEER_ADDR" // This optionally lets the current node dial a peer during boot up
	RaftDir  = "RAFT_DIR"
	HttpAddr = "HTTP_ADDR"
)
