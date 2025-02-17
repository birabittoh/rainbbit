package src

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const dataDir = "data"

var (
	db       *gorm.DB
	measures []string
)

// ------------------------
// MODELLI GORM
// ------------------------

// Record rappresenta i dati meteo completi (tranne la slice Weather, salvata separatamente)
type Record struct {
	Dt         int64 `json:"dt" gorm:"primarykey"`
	Visibility int   `json:"visibility"`

	// Sys
	Sunrise int64 `json:"sunrise"`
	Sunset  int64 `json:"sunset"`

	// Main
	Temp      float64 `json:"temp"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	FeelsLike float64 `json:"feels_like"`
	Pressure  float64 `json:"pressure"`
	SeaLevel  float64 `json:"sea_level"`
	GrndLevel float64 `json:"grnd_level"`
	Humidity  int     `json:"humidity"`

	// Wind
	WindSpeed float64 `json:"wind_speed"`
	WindDeg   float64 `json:"wind_deg"`

	// Clouds
	Clouds int `json:"clouds_all"`

	// Rain e Snow
	Rain1H float64 `json:"rain_1h" gorm:"column:rain_1h"`
	Snow1H float64 `json:"snow_1h" gorm:"column:snow_1h"`

	// ID numerici separati da ","
	Weather string `json:"weather"`

	Conditions []Condition `json:"conditions" gorm:"-"`
	Favicon    string      `json:"-" gorm:"-"`
	Title      string      `json:"-" gorm:"-"`
	TimeAgo    string      `json:"-" gorm:"-"`
}

func getAllRecords() (records []Record, err error) {
	err = db.Find(&records).Error
	if err != nil {
		return
	}

	for i := range records {
		records[i].parseConditions()
	}
	return
}

func getLatestRecord() (record Record, err error) {
	err = db.Last(&record).Error
	if err != nil {
		return
	}

	record.parseConditions()
	return
}

func initDB() (err error) {
	// Assicuriamoci che la directory "data" esista
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return errors.New("Errore nella creazione della directory 'data': " + err.Error())
	}

	// Inizializzazione del database SQLite con GORM
	dbPath := filepath.Join(dataDir, "data.sqlite?_pragma=foreign_keys(1)")
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return errors.New("Errore nell'apertura del database: " + err.Error())
	}

	// Migrazione dello schema per il modello Record
	if err := db.AutoMigrate(&Record{}); err != nil {
		return errors.New("Errore nella migrazione del database: " + err.Error())
	}

	// Inizializzazione delle colonne
	s, err := schema.Parse(&Record{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return errors.New("Errore nel parsing dello schema: " + err.Error())
	}
	for _, field := range s.Fields {
		if field.DBName == "" || field.DBName == "weather" || field.DBName == "dt" {
			continue
		}
		measures = append(measures, field.DBName)
	}

	return
}
