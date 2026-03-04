package utils

import "encoding/json"

func MarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
