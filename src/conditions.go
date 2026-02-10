package src

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	conditions map[string]Condition
	condOnce   sync.Once
	condMu     sync.RWMutex
	condErr    error
)

type Condition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func loadConditions() (map[string]Condition, error) {
	condOnce.Do(func() {
		file, err := os.Open("conditions.json")
		if err != nil {
			condErr = err
			return
		}
		defer file.Close()

		condMu.Lock()
		condErr = json.NewDecoder(file).Decode(&conditions)
		condMu.Unlock()
	})

	condMu.RLock()
	defer condMu.RUnlock()
	return conditions, condErr
}

func (record *Record) parseConditions() {
	dt := time.Unix(record.Dt, 0)
	sunrise := time.Unix(record.Sunrise, 0)
	sunset := time.Unix(record.Sunset, 0)

	weatherIDs := strings.Split(record.Weather, ",")

	condMu.RLock()
	defer condMu.RUnlock()
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
