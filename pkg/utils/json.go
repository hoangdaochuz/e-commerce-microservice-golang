package utils

import (
	"encoding/json"
)

func JsonStringToStruct[T any](val string) (*T, error) {
	var result T
	err := json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
