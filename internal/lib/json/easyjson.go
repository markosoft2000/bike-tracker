package json

import (
	"encoding/json"

	"github.com/mailru/easyjson"
)

func JSONEncoder(v any) ([]byte, error) {
	if marshaler, ok := v.(easyjson.Marshaler); ok {
		return easyjson.Marshal(marshaler)
	}

	return json.Marshal(v)
}

func JSONDecoder(data []byte, v any) error {
	if unmarshaler, ok := v.(easyjson.Unmarshaler); ok {
		return easyjson.Unmarshal(data, unmarshaler)
	}

	return json.Unmarshal(data, v)
}
