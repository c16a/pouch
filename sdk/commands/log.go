package commands

import (
	"errors"
	"fmt"
)

type JoinResponse struct {
	OK  bool  `json:"ok"`
	Err error `json:"err"`
}

func (r JoinResponse) String() string {
	if r.Err != nil {
		return fmt.Sprintf("ERR %s", r.Err.Error())
	}
	return fmt.Sprintf("OK %v", r.OK)
}

var (
	ErrorInvalidDataType = errors.New("invalid data type")
	ErrorNotFound        = errors.New("not found")
)

type ErrorResponse struct {
	Err error
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("ERR %s", e.Err.Error())
}

type CountResponse struct {
	Count int
}

func (c *CountResponse) String() string {
	return fmt.Sprintf("COUNT %d", c.Count)
}

type NilResponse struct {
}

func (n *NilResponse) String() string {
	return "NIL"
}

type BooleanResponse struct {
	Value bool
}

func (b *BooleanResponse) String() string {
	return fmt.Sprintf("BOOLEAN %v", b.Value)
}

type StringResponse struct {
	Value string
}

func (s *StringResponse) String() string {
	return fmt.Sprintf("STRING %s", s.Value)
}

type ListResponse struct {
	Values []string
}

func (l *ListResponse) String() string {
	var result string
	for i, value := range l.Values {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("(%d): %s", i, value)
	}
	return result
}
