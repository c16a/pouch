package commands

import (
	"strings"
)

// MessageType is a request from client
type MessageType string

const (
	Join MessageType = "JOIN"
	Get  MessageType = "GET"
	Set  MessageType = "SET"
	Del  MessageType = "DEL"

	LPush  MessageType = "LPUSH"
	RPush  MessageType = "RPUSH"
	LPop   MessageType = "LPOP"
	RPop   MessageType = "RPOP"
	LRange MessageType = "LRANGE"
	LLen   MessageType = "LLEN"

	SAdd      MessageType = "SADD"
	SCard     MessageType = "SCARD"
	SDiff     MessageType = "SDIFF"
	SInter    MessageType = "SINTER"
	SIsMember MessageType = "SISMEMBER"
	SMembers  MessageType = "SMEMBERS"
	SUnion    MessageType = "SUNION"

	AuthChallengeResponse MessageType = "AUTH.CHALLENGE.RES"
	AuthChallengeRequest  MessageType = "AUTH.CHALLENGE.REQ"

	Err     MessageType = "ERR"
	Count   MessageType = "COUNT"
	String  MessageType = "STRING"
	Boolean MessageType = "BOOLEAN"
)

type Command interface {
	GetAction() MessageType
	String() string
}

func ParseStringIntoCommand(s string) (Command, error) {
	parts := strings.Split(s, " ")

	if len(parts) == 0 {
		return nil, ErrEmptyCommand
	}

	action := parts[0]

	lineMessage := LineMessage{Line: s, MessageType: MessageType(action)}

	switch action {
	// This is a special action for joining clusters
	case string(Join):
		return &JoinCommand{NodeId: parts[1], Addr: parts[2]}, nil
	case string(AuthChallengeResponse):
		return NewAuthChallengeResponseCommand(lineMessage)
	case string(AuthChallengeRequest):
		return NewAuthChallengeRequestCommand(lineMessage)
	case string(Get):
		return NewGetCommand(lineMessage)
	case string(Set):
		return NewSetCommand(lineMessage)
	case string(Del):
		return NewDelCommand(lineMessage)
	case string(LPush):
		return NewLPushCommand(lineMessage)
	case string(RPush):
		return NewRPushCommand(lineMessage)
	case string(LLen):
		return NewLLenCommand(lineMessage)
	case string(LPop):
		return NewLPopCommand(lineMessage)
	case string(RPop):
		return NewRPopCommand(lineMessage)
	case string(LRange):
		return NewLRangeCommand(lineMessage)
	case string(SAdd):
		return NewSAddCommand(lineMessage)
	case string(SCard):
		return NewSCardCommand(lineMessage)
	case string(SDiff):
		return NewSDiffCommand(lineMessage)
	case string(SInter):
		return NewSInterCommand(lineMessage)
	case string(SUnion):
		return NewSUnionCommand(lineMessage)
	case string(SIsMember):
		return NewSIsMemberCommand(lineMessage)
	case string(SMembers):
		return NewSMembersCommand(lineMessage)
	default:
		return nil, ErrInvalidCommand
	}
}
