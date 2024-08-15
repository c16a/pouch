package commands

import (
	"errors"
	"strconv"
	"strings"
)

type CommandKind string

// CommandAction is a request from client
type CommandAction string

const (
	Join CommandAction = "JOIN"
	Get  CommandAction = "GET"
	Set  CommandAction = "SET"
	Del  CommandAction = "DEL"

	LPush  CommandAction = "LPUSH"
	RPush  CommandAction = "RPUSH"
	LPop   CommandAction = "LPOP"
	RPop   CommandAction = "RPOP"
	LRange CommandAction = "LRANGE"
	LLen   CommandAction = "LLEN"

	SAdd      CommandAction = "SADD"
	SCard     CommandAction = "SCARD"
	SDiff     CommandAction = "SDIFF"
	SInter    CommandAction = "SINTER"
	SIsMember CommandAction = "SISMEMBER"
	SMembers  CommandAction = "SMEMBERS"
	SUnion    CommandAction = "SUNION"
)

type Command interface {
	GetAction() CommandAction
	String() string
}

func ParseStringIntoCommand(s string) (Command, error) {
	parts := strings.Split(s, " ")

	if len(parts) == 0 {
		return nil, errors.New("invalid command")
	}

	action := parts[0]

	switch action {
	// This is a special action for joining clusters
	case string(Join):
		return &JoinCommand{NodeId: parts[1], Addr: parts[2]}, nil

	case string(Get):
		return &GetCommand{Key: parts[1], line: s}, nil
	case string(Set):
		return &SetCommand{Key: parts[1], Value: parts[2], line: s}, nil
	case string(Del):
		return &DelCommand{Key: parts[1], line: s}, nil
	case string(LPush):
		return &LPushCommand{Key: parts[1], Values: parts[2:], line: s}, nil
	case string(RPush):
		return &RPushCommand{Key: parts[1], Values: parts[2:], line: s}, nil
	case string(LLen):
		return &LLenCommand{Key: parts[1], line: s}, nil
	case string(LPop):
		var err error
		key := parts[1]

		count := 1
		if len(parts) == 3 {
			count, err = strconv.Atoi(parts[2])
			if err != nil {
				return nil, errors.New("invalid count")
			}
		}
		return &LPopCommand{Key: key, Count: count, line: s}, nil
	case string(RPop):
		var err error
		key := parts[1]

		count := 1
		if len(parts) == 3 {
			count, err = strconv.Atoi(parts[2])
			if err != nil {
				return nil, errors.New("invalid count")
			}
		}

		return &RPopCommand{Key: key, Count: count, line: s}, nil
	case string(LRange):
		var err error

		key := parts[1]

		startIdx := 0
		if len(parts) == 3 {
			startIdx, err = strconv.Atoi(parts[2])
			if err != nil {
				return nil, errors.New("invalid start index")
			}
		}

		endIdx := -1
		if len(parts) == 4 {
			endIdx, err = strconv.Atoi(parts[3])
			if err != nil {
				return nil, errors.New("invalid end index")
			}
		}
		return &LRangeCommand{Key: key, Start: startIdx, End: endIdx, line: s}, nil
	case string(SAdd):
		return &SAddCommand{Key: parts[1], Values: parts[2:], line: s}, nil
	case string(SCard):
		return &SCardCommand{Key: parts[1], line: s}, nil
	case string(SDiff):
		return &SDiffCommand{Key: parts[1], OtherKeys: parts[2:], line: s}, nil
	case string(SInter):
		return &SInterCommand{Key: parts[1], OtherKeys: parts[2:], line: s}, nil
	case string(SUnion):
		return &SUnionCommand{Key: parts[1], OtherKeys: parts[2:], line: s}, nil
	case string(SIsMember):
		return &SIsMemberCommand{Key: parts[1], Value: parts[2], line: s}, nil
	case string(SMembers):
		return &SMembersCommand{Key: parts[1], line: s}, nil
	default:
		return nil, errors.New("invalid command")
	}
}
