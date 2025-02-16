package src

import (
	"log"
	"strings"
	"time"

	"github.com/briandowns/openweathermap"
	"gorm.io/gorm"
)

// ------------------------
// FUNZIONI DI SUPPORTO
// ------------------------

// fetchAndSaveWeather effettua la chiamata all'API, mappa i dati nei modelli e li salva nel database.
func fetchAndSaveWeather(db *gorm.DB, coords *openweathermap.Coordinates, apiKey, unit, lang string) {
	// Creazione dell'oggetto per il meteo corrente
	current, err := openweathermap.NewCurrent(unit, lang, apiKey)
	if err != nil {
		log.Println("Errore nella creazione dell'oggetto OpenWeatherMap:", err)
		return
	}

	// Chiamata all'API usando le coordinate specificate
	err = current.CurrentByCoordinates(coords)
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
		// Wind
		WindSpeed: current.Wind.Speed,
		WindDeg:   current.Wind.Deg,
		// Clouds
		Clouds: current.Clouds.All,
		// Rain e Snow salvati come stringa JSON (potranno essere vuoti "{}")
		Rain1H: current.Rain.OneH,
		Snow1H: current.Snow.OneH,
	}

	// Per ogni elemento dell'array Weather, creiamo un oggetto Weather
	for _, w := range current.Weather {
		wr := Weather{WeatherID: w.ID}
		record.Weather = append(record.Weather, wr)
	}

	// Salvataggio nel database (includendo l'associazione Weathers)
	if err := db.Create(&record).Error; err != nil {
		log.Println("Errore nel salvataggio del record:", err)
		return
	}
	log.Printf("Record salvato alle %s\n", time.Now().Format(time.RFC3339))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
