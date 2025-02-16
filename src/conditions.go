package src

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var conditions map[int]Condition

type Condition struct {
	ID          int    `json:"id"`
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

func (record *Record) parseConditions() {
	dt := time.Unix(record.Dt, 0)
	sunrise := time.Unix(record.Sunrise, 0)
	sunset := time.Unix(record.Sunset, 0)

	for _, w := range record.Weather {
		c, ok := conditions[w.WeatherID]
		if !ok {
			log.Printf("Condizione meteo non trovata per ID %d\n", w.WeatherID)
			continue
		}

		if dt.After(sunrise) && dt.Before(sunset) {
			c.Icon += "d"
		} else {
			c.Icon += "n"
		}

		c.ID = w.WeatherID
		c.Description = capitalize(c.Description)

		record.Conditions = append(record.Conditions, c)
	}

	record.Favicon = record.Conditions[0].Icon
	record.Title = record.Conditions[0].Description
	record.TimeAgo = time.Since(dt).Round(time.Second).String()
}
