package store

type NodeConfig struct {
	Tcp      *Tcp
	Ws       *Ws
	Quic     *Quic
	Unix     *Unix
	Auth     *Auth
	Cluster  *Cluster
	Security *Security
}

type Tcp struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"`
}

type Quic struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"`
}

type Unix struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

type Ws struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr"`
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

type Security struct {
	Tls *Tls `json:"tls"`
}

type Tls struct {
	Enable       bool   `json:"enable"`
	CertFilePath string `json:"cert_file_path"`
	KeyFilePath  string `json:"key_file_path"`
}
