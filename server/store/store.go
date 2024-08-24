package store

import (
	"time"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)
