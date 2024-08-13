package commands

import "fmt"

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

type ErrorResponse struct {
	Err error
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("ERR %s", e.Err.Error())
}

type CountResponse struct {
	Count int `json:"COUNT"`
}

func (c *CountResponse) String() string {
	return fmt.Sprintf("COUNT %d", c.Count)
}

type NilResponse struct {
}

func (n *NilResponse) String() string {
	return "NIL"
}

type StringResponse struct {
	Value string
}

func (s *StringResponse) String() string {
	return fmt.Sprintf("STRING %s", s.Value)
}
