package utils

import (
	"encoding/json"
	"fmt"
)

func MapToStruct(m map[string]interface{}, s interface{}) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshaling map: %w", err)
	}

	if err = json.Unmarshal(jsonData, s); err != nil {
		return fmt.Errorf("error unmarshaling map: %w", err)
	}

	return nil
}
