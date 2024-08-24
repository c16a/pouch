package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type LineMessage struct {
	Line        string
	MessageType MessageType
}

func (l *LineMessage) String() string {
	return l.Line
}

func (l *LineMessage) GetMessageType() MessageType {
	return l.MessageType
}

type DelCommand struct {
	Key string
	LineMessage
}

func NewDelCommand(line LineMessage) (*DelCommand, error) {
	parts := strings.Split(line.Line, " ")
	return &DelCommand{Key: parts[1], LineMessage: line}, nil
}

type LPushCommand struct {
	Key    string
	Values []string
	LineMessage
}

func NewLPushCommand(line LineMessage) (*LPushCommand, error) {
	parts := strings.Split(line.Line, " ")
	return &LPushCommand{
		Key:         parts[1],
		Values:      parts[2:],
		LineMessage: line,
	}, nil
}

type RPushCommand struct {
	Key    string
	Values []string
	LineMessage
}

func NewRPushCommand(line LineMessage) (*RPushCommand, error) {
	parts := strings.Split(line.Line, " ")
	return &RPushCommand{
		Key:         parts[1],
		Values:      parts[2:],
		LineMessage: line,
	}, nil
}

type LLenCommand struct {
	Key string
	LineMessage
}

func NewLLenCommand(line LineMessage) (*LLenCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &LLenCommand{
		Key:         parts[1],
		LineMessage: line,
	}, nil
}

type LPopCommand struct {
	Key   string
	Count int
	LineMessage
}

func NewLPopCommand(line LineMessage) (*LPopCommand, error) {
	parts := strings.Split(line.String(), " ")

	var err error
	key := parts[1]

	count := 1
	if len(parts) == 3 {
		count, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, errors.New("invalid count")
		}
	}
	return &LPopCommand{Key: key, Count: count, LineMessage: line}, nil
}

type RPopCommand struct {
	Key   string
	Count int
	LineMessage
}

func NewRPopCommand(line LineMessage) (*RPopCommand, error) {
	parts := strings.Split(line.String(), " ")
	var err error
	key := parts[1]

	count := 1
	if len(parts) == 3 {
		count, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, errors.New("invalid count")
		}
	}

	return &RPopCommand{Key: key, Count: count, LineMessage: line}, nil
}

type LRangeCommand struct {
	Key   string
	Start int
	End   int
	LineMessage
}

func NewLRangeCommand(line LineMessage) (*LRangeCommand, error) {
	parts := strings.Split(line.String(), " ")
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
	return &LRangeCommand{Key: key, Start: startIdx, End: endIdx, LineMessage: line}, nil
}

type SAddCommand struct {
	Key    string
	Values []string
	LineMessage
}

func NewSAddCommand(line LineMessage) (*SAddCommand, error) {
	parts := strings.Split(line.String(), ":")
	return &SAddCommand{
		Key:         parts[1],
		Values:      parts[2:],
		LineMessage: line,
	}, nil
}

type SCardCommand struct {
	Key string
	LineMessage
}

func NewSCardCommand(line LineMessage) (*SCardCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SCardCommand{
		Key:         parts[1],
		LineMessage: line,
	}, nil
}

type SDiffCommand struct {
	Key       string
	OtherKeys []string
	LineMessage
}

func NewSDiffCommand(line LineMessage) (*SDiffCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SDiffCommand{
		Key:         parts[1],
		OtherKeys:   parts[2:],
		LineMessage: line,
	}, nil
}

type SInterCommand struct {
	Key       string
	OtherKeys []string
	LineMessage
}

func NewSInterCommand(line LineMessage) (*SInterCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SInterCommand{
		Key:         parts[1],
		OtherKeys:   parts[2:],
		LineMessage: line,
	}, nil
}

type SUnionCommand struct {
	Key       string
	OtherKeys []string
	LineMessage
}

func NewSUnionCommand(line LineMessage) (*SUnionCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SUnionCommand{
		Key:         parts[1],
		OtherKeys:   parts[2:],
		LineMessage: line,
	}, nil
}

type SIsMemberCommand struct {
	Key   string
	Value string
	LineMessage
}

func NewSIsMemberCommand(line LineMessage) (*SIsMemberCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SIsMemberCommand{
		LineMessage: line,
		Key:         parts[1],
		Value:       parts[2],
	}, nil
}

type SMembersCommand struct {
	Key string
	LineMessage
}

func NewSMembersCommand(line LineMessage) (*SMembersCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SMembersCommand{
		Key:         parts[1],
		LineMessage: line,
	}, nil
}

type PFAddCommand struct {
	Key    string
	Values []string
	LineMessage
}

func NewPFAddCommand(line LineMessage) (*PFAddCommand, error) {
	parts := strings.Split(line.String(), ":")
	return &PFAddCommand{
		Key:         parts[1],
		Values:      parts[2:],
		LineMessage: line,
	}, nil
}

type PFCountCommand struct {
	Key string
	LineMessage
}

func NewPFCountCommand(line LineMessage) (*PFCountCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &PFCountCommand{
		Key:         parts[1],
		LineMessage: line,
	}, nil
}

type PFMergeCommand struct {
	DestKey    string
	SourceKeys []string
	LineMessage
}

func NewPFMergeCommand(line LineMessage) (*PFMergeCommand, error) {
	parts := strings.Split(line.String(), ":")
	return &PFMergeCommand{
		DestKey:     parts[1],
		SourceKeys:  parts[2:],
		LineMessage: line,
	}, nil
}

type GetCommand struct {
	Key string
	LineMessage
}

func NewGetCommand(line LineMessage) (*GetCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &GetCommand{
		LineMessage: line,
		Key:         parts[1],
	}, nil
}

type SetCommand struct {
	Key   string
	Value string
	LineMessage
}

func NewSetCommand(line LineMessage) (*SetCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &SetCommand{
		LineMessage: line,
		Key:         parts[1],
		Value:       parts[2],
	}, nil
}

type AuthChallengeResponseCommand struct {
	ClientId           string
	ChallengeSignature string
	LineMessage
}

func NewAuthChallengeResponseCommand(line LineMessage) (*AuthChallengeResponseCommand, error) {
	parts := strings.Split(line.String(), " ")
	return &AuthChallengeResponseCommand{
		ClientId:           parts[1],
		ChallengeSignature: parts[2],
		LineMessage:        line,
	}, nil
}

func NewAuthChallengeResponseCommandWithValues(client string, challengeSignature string) (*AuthChallengeResponseCommand, error) {
	line := LineMessage{
		Line:        fmt.Sprintf("%s %s %s", AuthChallengeResponse, client, challengeSignature),
		MessageType: AuthChallengeResponse,
	}
	return NewAuthChallengeResponseCommand(line)
}
