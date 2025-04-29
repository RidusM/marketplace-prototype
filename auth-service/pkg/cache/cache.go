package cache

import (
	"encoding/json"
	"fmt"
)

func GenerateCacheKey(prefix string, params any) string {
	return fmt.Sprintf("%s:%v", prefix, params)
}

func Serialize(data any) ([]byte, error) {
	res, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("cache.Serialize: %w", err)
	}

	return res, nil
}

func Deserialize(data []byte, output any) error {
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("cache.Deserialize: %w", err)
	}

	return nil
}
