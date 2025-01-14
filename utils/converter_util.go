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

	err = json.Unmarshal(jsonData, s)
	if err != nil {
		return fmt.Errorf("error unmarshaling to struct: %w", err)
	}

	return nil
}
