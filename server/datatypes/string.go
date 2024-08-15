package datatypes

import "encoding/json"

type String struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func NewString(s string) *String {
	return &String{Value: s, Name: "string"}
}

func (s *String) GetName() string {
	return s.Name
}

func (s *String) GetValue() string {
	return s.Value
}
