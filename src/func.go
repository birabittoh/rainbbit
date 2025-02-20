package src

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/briandowns/openweathermap"
	"gorm.io/gorm"
)

var current *openweathermap.CurrentWeatherData

// ------------------------
// FUNZIONI DI SUPPORTO
// ------------------------

// fetchAndSaveWeather effettua la chiamata all'API, mappa i dati nei modelli e li salva nel database.
func fetchAndSaveWeather(db *gorm.DB, coords *openweathermap.Coordinates) {
	// Chiamata all'API usando le coordinate specificate
	err := current.CurrentByCoordinates(coords)
	if err != nil {
		log.Println("Errore nella chiamata API:", err)
		return
	}

	// Mappatura dei dati restituiti nel modello Record
	record := Record{
		Dt:         int64(current.Dt),
		Visibility: current.Visibility,
		// Sys
		Sunrise: int64(current.Sys.Sunrise),
		Sunset:  int64(current.Sys.Sunset),
		// Main
		Temp:      current.Main.Temp,
		TempMin:   current.Main.TempMin,
		TempMax:   current.Main.TempMax,
		FeelsLike: current.Main.FeelsLike,
		Pressure:  current.Main.Pressure,
		SeaLevel:  current.Main.SeaLevel,
		GrndLevel: current.Main.GrndLevel,
		Humidity:  current.Main.Humidity,
		// Other
		WindSpeed: current.Wind.Speed,
		WindDeg:   current.Wind.Deg,
		Clouds:    current.Clouds.All,
		Rain1H:    current.Rain.OneH,
		Snow1H:    current.Snow.OneH,
	}

	weatherIDs := []string{}
	for _, w := range current.Weather {
		weatherIDs = append(weatherIDs, strconv.Itoa(w.ID))
	}
	record.Weather = strings.Join(weatherIDs, ",")

	// Salvataggio nel database (includendo l'associazione Weathers)
	if err := db.Create(&record).Error; err != nil {
		log.Println("Errore nel salvataggio del record:", err)
		return
	}
	log.Println("Record salvato")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func getHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func getEnvDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}
