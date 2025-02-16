package src

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/briandowns/openweathermap"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Impostazioni per unità e lingua
const (
	unit = "C"
	lang = "it"
	port = "3000"
)

var db *gorm.DB

// ------------------------
// FUNZIONE MAIN
// ------------------------
func Main() {
	// Caricamento delle variabili d'ambiente dal file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Nessun file .env trovato, verranno usate le variabili d'ambiente di sistema")
	}

	// Lettura delle variabili d'ambiente necessarie
	apiKey := os.Getenv("OWM_API_KEY")
	latitudeStr := os.Getenv("OWM_LATITUDE")
	longitudeStr := os.Getenv("OWM_LONGITUDE")

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		log.Fatalln("Errore nel parsing di OWM_LATITUDE:", err)
	}
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		log.Fatalln("Errore nel parsing di OWM_LONGITUDE:", err)
	}

	coords := &openweathermap.Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}

	// Assicuriamoci che la directory "data" esista
	dataDir := "data"
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Fatalln("Errore nella creazione della directory 'data':", err)
	}

	// Inizializzazione del database SQLite con GORM
	dbPath := filepath.Join(dataDir, "data.sqlite?_pragma=foreign_keys(1)")
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalln("Errore nell'apertura del database:", err)
	}

	// Migrazione degli schemi per i modelli Record e WeatherRecord
	if err := db.AutoMigrate(&Record{}, &Weather{}); err != nil {
		log.Fatalln("Errore nella migrazione del database:", err)
	}

	// Creazione e configurazione del cron scheduler.
	// In questo esempio il job viene eseguito ogni mezz'ora (minuti 0 e 30)
	c := cron.New(cron.WithSeconds())
	spec := "0 0/30 * * * *"
	_, err = c.AddFunc(spec, func() {
		log.Println("Esecuzione fetchAndSaveWeather:", time.Now().Format(time.RFC3339))
		fetchAndSaveWeather(db, coords, apiKey, unit, lang)
	})
	if err != nil {
		log.Fatalln("Errore nella creazione del cron job:", err)
	}

	// Avvio del cron scheduler
	c.Start()
	log.Println("Scheduler avviato, in attesa di esecuzioni...")

	// Blocca il main per mantenere il programma in esecuzione
	mux := getServeMux()
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
