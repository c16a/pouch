package store

import (
	"encoding/json"
	"github.com/c16a/pouch/server/datatypes"
	"github.com/hashicorp/raft"
)

type FsmSnapshot struct {
	store map[string]datatypes.Type
}

func (f *FsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *FsmSnapshot) Release() {}
