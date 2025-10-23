package utils

import (
	"encoding/json"

	"golang.org/x/text/encoding"
)

func StructToJsonString(val any, encoders ...encoding.Encoder) (string, error) {
	byteStruct, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	if len(encoders) > 0 {
		encodedByte, err := encoders[0].Bytes(byteStruct)
		if err != nil {
			return "", err
		}
		return string(encodedByte), nil
	}
	return string(byteStruct), nil
}
