package utils

import (
	"encoding/json"
	"fmt"
)

func GenerateCacheKey(prefix string, params any) string {
	return fmt.Sprintf("%s:%v", prefix, params)
}

func GenerateCacheKeyParams(params ...any) string {
	var str string

	for i, param := range params {
		str += fmt.Sprintf("%v", param)

		last := len(params) - 1
		if i != last {
			str += "-"
		}
	}

	return str
}

func Serialize(data any) ([]byte, error) {
	return json.Marshal(data)
}

func Deserialize(data []byte, output any) error {
	return json.Unmarshal(data, output)
}
