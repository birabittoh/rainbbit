package src

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

var conditions map[string]Condition

type Condition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func loadConditions() (map[string]Condition, error) {
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

func (record *Record) parseConditions() {
	dt := time.Unix(record.Dt, 0)
	sunrise := time.Unix(record.Sunrise, 0)
	sunset := time.Unix(record.Sunset, 0)

	weatherIDs := strings.Split(record.Weather, ",")

	for _, w := range weatherIDs {
		c, ok := conditions[w]
		if !ok {
			log.Printf("Condizione meteo non trovata per ID %s\n", w)
			continue
		}

		if dt.After(sunrise) && dt.Before(sunset) {
			c.Icon += "d"
		} else {
			c.Icon += "n"
		}

		c.Description = strings.Split(c.Description, ": ")[0]

		record.Conditions = append(record.Conditions, c)
	}
}
