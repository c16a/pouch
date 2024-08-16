package store

import (
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/datatypes"
)

func (node *RaftNode) LLen(cmd *commands.LLenCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			response := &commands.CountResponse{Count: listVal.LLen()}
			return response.String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *RaftNode) LRange(cmd *commands.LRangeCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			lrange, err := listVal.LRange(cmd.Start, cmd.End)
			if err != nil {
				return (&commands.ErrorResponse{Err: err}).String()
			}
			response := &commands.ListResponse{Values: lrange}
			return response.String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *RaftNode) LPush(cmd *commands.LPushCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) RPush(cmd *commands.RPushCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) RPop(cmd *commands.RPopCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) LPop(cmd *commands.LPopCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *RaftNode) applyLPush(cmd *commands.LPushCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			listVal.LPushAll(cmd.Values)
			return (&commands.CountResponse{Count: len(cmd.Values)}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		list := datatypes.NewList()
		list.LPushAll(cmd.Values)
		node.m[cmd.Key] = list
		return (&commands.CountResponse{Count: len(cmd.Values)}).String()
	}
}

func (node *RaftNode) applyRPush(cmd *commands.RPushCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			listVal.RPushAll(cmd.Values)
			return (&commands.CountResponse{Count: len(cmd.Values)}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		list := datatypes.NewList()
		list.RPushAll(cmd.Values)
		node.m[cmd.Key] = list
		return (&commands.CountResponse{Count: len(cmd.Values)}).String()
	}
}

func (node *RaftNode) applyLpop(cmd *commands.LPopCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			if res, err := listVal.LPopN(cmd.Count); err == nil {
				return (&commands.ListResponse{Values: res}).String()
			} else {
				return (&commands.ErrorResponse{Err: err}).String()
			}
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *RaftNode) applyRpop(cmd *commands.RPopCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "list":
			listVal := val.(*datatypes.List)
			if res, err := listVal.RPopN(cmd.Count); err == nil {
				return (&commands.ListResponse{Values: res}).String()
			} else {
				return (&commands.ErrorResponse{Err: err}).String()
			}
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}
