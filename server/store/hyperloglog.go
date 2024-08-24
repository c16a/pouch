package store

import (
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/datatypes"
)

func (node *RaftNode) PFAdd(cmd *commands.PFAddCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) PFCount(cmd *commands.PFCountCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "hll":
			hll := val.(*datatypes.HyperLogLog)
			return (&commands.CountResponse{Count: int(hll.Estimate())}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *RaftNode) applyPFAdd(cmd *commands.PFAddCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "hll":
			hll := val.(*datatypes.HyperLogLog)
			count := hll.AddMany(cmd.Values)
			return (&commands.CountResponse{Count: count}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		hll := datatypes.NewHllWithErrorRate(0.6)
		count := hll.AddMany(cmd.Values)
		node.m[cmd.Key] = hll
		return (&commands.CountResponse{Count: count}).String()
	}
}
