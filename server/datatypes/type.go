package datatypes

import "encoding/json"

type Type interface {
	GetName() string
	json.Marshaler
}
