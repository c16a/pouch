package store

type NodeConfig struct {
	Tcp     *Tcp
	Ws      *Ws
	Unix    *Unix
	Auth    *Auth
	Cluster *Cluster
}

type Tcp struct {
	Enabled bool   `json:"enable"`
	Addr    string `json:"addr"`
}

type Unix struct {
	Enabled bool   `json:"enable"`
	Path    string `json:"path"`
}

type Ws struct {
	Enable bool   `json:"enable"`
	Addr   string `json:"addr"`
}

type Cluster struct {
	NodeID    string   `json:"node_id"`
	Addr      string   `json:"addr"`
	RaftDir   string   `json:"raft_dir"`
	PeerAddrs []string `json:"peer_addrs"`
}

type Auth struct {
	Clients map[string]*ClientInfo
}

type ClientInfo struct {
	HexPublicKey string `json:"hex_public_key"`
}
