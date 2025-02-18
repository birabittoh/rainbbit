package src

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/briandowns/openweathermap"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

// Impostazioni per unità e lingua
const (
	unit    = "C"
	lang    = "it"
	address = ":3000"
)

// ------------------------
// FUNZIONE MAIN
// ------------------------
func Main() {
	// Caricamento delle variabili d'ambiente dal file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Nessun file .env trovato, verranno usate le variabili d'ambiente di sistema")
	}

	// Lettura delle variabili d'ambiente necessarie
	apiKey := os.Getenv("OWM_API_KEY")
	latitudeStr := os.Getenv("OWM_LATITUDE")
	longitudeStr := os.Getenv("OWM_LONGITUDE")

	coords := &openweathermap.Coordinates{}
	coords.Latitude, err = strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		log.Fatalln("Errore nel parsing di OWM_LATITUDE:", err)
	}
	coords.Longitude, err = strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		log.Fatalln("Errore nel parsing di OWM_LONGITUDE:", err)
	}

	// Connessione al database
	err = initDB()
	if err != nil {
		log.Fatalln("Errore nell'inizializzazione del database:", err)
	}

	// Inizializzazione di OpenWeatherMap
	current, err = openweathermap.NewCurrent(unit, lang, apiKey)
	if err != nil {
		log.Fatal("Errore nella creazione dell'oggetto OpenWeatherMap:", err)
		return
	}

	// Creazione e configurazione del cron scheduler
	spec := getEnvDefault("OWM_CRON", "0 0/30 * * * *")
	c := cron.New(cron.WithSeconds())
	_, err = c.AddFunc(spec, func() {
		log.Println("Eseguo fetchAndSaveWeather")
		fetchAndSaveWeather(db, coords)
	})
	if err != nil {
		log.Fatalln("Errore nella creazione del cron job:", err)
	}

	// Avvio del cron scheduler
	c.Start()
	log.Println("Cron scheduler avviato")

	// Aggiungo un primo record nel database, se necessario
	var count int64
	err = db.Model(&Record{}).Count(&count).Error
	if err != nil {
		log.Fatal("Errore durante il controllo dei record nel database:", err)
	}
	if count == 0 {
		log.Println("Nessun record trovato nel database, eseguo fetchAndSaveWeather")
		fetchAndSaveWeather(db, coords)
	}

	address := getEnvDefault("APP_ADDRESS", ":3000")
	// Avvio del server HTTP
	s := &http.Server{
		Addr:              address,
		Handler:           rateLimiterMiddleware(getServeMux()),
		ReadTimeout:       5 * time.Second,  // Timeout per leggere la richiesta
		WriteTimeout:      10 * time.Second, // Timeout per scrivere la risposta
		IdleTimeout:       60 * time.Second, // Timeout per connessioni Keep-Alive
		ReadHeaderTimeout: 2 * time.Second,  // Previene attacchi Slowloris
	}

	log.Println("Server in ascolto all'indirizzo", address)
	log.Fatal(s.ListenAndServe())
}
