package gormdb

import (
	"encoding/json"
)

func encodeImageSlice(images []string) (string, error) {
	if images == nil {
		images = []string{}
	}
	b, err := json.Marshal(images)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeImageSlice(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}
