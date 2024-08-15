package store

import (
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/datatypes"
)

func (node *Node) SAdd(cmd *commands.SAddCommand) string {
	return node.respondAfterRaftCommit(cmd)
}

func (node *Node) applySADD(cmd *commands.SAddCommand) interface{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "set":
			setVal := val.(*datatypes.Set[string])
			count := setVal.AddMany(cmd.Values)
			return (&commands.CountResponse{Count: count}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		set := datatypes.NewSet[string]()
		count := set.AddMany(cmd.Values)
		node.m[cmd.Key] = set
		return (&commands.CountResponse{Count: count}).String()
	}
}

func (node *Node) SCard(cmd *commands.SCardCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "set":
			setVal := val.(*datatypes.Set[string])
			return (&commands.CountResponse{Count: setVal.Size()}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *Node) SMembers(cmd *commands.SMembersCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "set":
			setVal := val.(*datatypes.Set[string])
			return (&commands.ListResponse{Values: setVal.GetMembers()}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *Node) SIsMember(cmd *commands.SIsMemberCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()

	if val, ok := node.m[cmd.Key]; ok {
		switch val.GetName() {
		case "set":
			setVal := val.(*datatypes.Set[string])
			found := setVal.Contains(cmd.Value)
			return (&commands.BooleanResponse{Value: found}).String()
		default:
			return (&commands.ErrorResponse{Err: commands.ErrorInvalidDataType}).String()
		}
	} else {
		return (&commands.ErrorResponse{Err: commands.ErrorNotFound}).String()
	}
}

func (node *Node) SUnion(cmd *commands.SUnionCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()

	set, err := node.findSet(cmd.Key)
	if err != nil {
		return (&commands.ErrorResponse{Err: err}).String()
	}

	union := set.Copy()
	for _, otherKey := range cmd.OtherKeys {
		otherSet, err := node.findSet(otherKey)
		if err != nil {
			continue
		}
		union = union.Union(otherSet)
	}

	return (&commands.ListResponse{Values: union.GetMembers()}).String()
}

func (node *Node) SInter(cmd *commands.SInterCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()

	set, err := node.findSet(cmd.Key)
	if err != nil {
		return (&commands.ErrorResponse{Err: err}).String()
	}

	intersection := set.Copy()
	for _, otherKey := range cmd.OtherKeys {
		otherSet, err := node.findSet(otherKey)
		if err != nil {
			continue
		}
		intersection = intersection.Intersection(otherSet)
	}

	return (&commands.ListResponse{Values: intersection.GetMembers()}).String()
}

func (node *Node) SDiff(cmd *commands.SDiffCommand) string {
	node.mu.Lock()
	defer node.mu.Unlock()

	set, err := node.findSet(cmd.Key)
	if err != nil {
		return (&commands.ErrorResponse{Err: err}).String()
	}

	diff := set.Copy()
	for _, otherKey := range cmd.OtherKeys {
		otherSet, err := node.findSet(otherKey)
		if err != nil {
			continue
		}
		diff = diff.Difference(otherSet)
	}

	return (&commands.ListResponse{Values: diff.GetMembers()}).String()
}

func (node *Node) findSet(key string) (*datatypes.Set[string], error) {
	if val, ok := node.m[key]; ok {
		switch val.GetName() {
		case "set":
			setVal := val.(*datatypes.Set[string])
			return setVal, nil
		default:
			return nil, commands.ErrorInvalidDataType
		}
	} else {
		return nil, commands.ErrorNotFound
	}
}
