package utils

import "encoding/json"

// StructToMap converts a struct to a map using json marshalling
func StructToMap(v interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if v != nil {
		jsonString, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonString, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
