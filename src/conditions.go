package src

import (
	"encoding/json"
	"os"
)

var conditions map[int]Condition

type Condition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func loadConditions() (map[int]Condition, error) {
	file, err := os.Open("conditions.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&conditions)
	if err != nil {
		return nil, err
	}

	return conditions, nil
}
