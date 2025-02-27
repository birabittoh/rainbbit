package src

import (
	"errors"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/birabittoh/myks"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	dataDir   = "data"
	dbPath    = dataDir + string(os.PathSeparator) + "data.sqlite"
	dbOptions = "?_pragma=foreign_keys(1)"
	zonePath  = dataDir + string(os.PathSeparator) + "zone.txt"
)

var (
	db       *gorm.DB
	measures []string

	recordsCache = myks.New[[]Record](30 * time.Minute)
	dpCache      = myks.New[[]DataPoint](31 * time.Minute)
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
}

func alignConstraints(from int64, to int64) (f, t *int64) {
	// round down to the nearest cron interval
	alignedFrom := (from / cronInterval) * cronInterval
	// round up to the nearest cron interval
	alignedTo := ((to + cronInterval - 1) / cronInterval) * cronInterval

	return &alignedFrom, &alignedTo
}

func addConstraints(query *gorm.DB, from, to *int64) *gorm.DB {
	if from != nil {
		query = query.Where("dt >= ?", *from)
	}
	if to != nil {
		query = query.Where("dt <= ?", *to)
	}
	return query
}

func getAllRecords(from int64, to int64) (records []Record, err error) {
	f, t := alignConstraints(from, to)
	key := getKey([]string{"*"}, f, t)

	value, err := recordsCache.Get(key)
	if err == nil {
		records = *value
		return
	}

	query := addConstraints(db.Model(&Record{}), f, t)
	err = query.Find(&records).Error
	if err != nil {
		return
	}

	for i := range records {
		records[i].parseConditions()
	}

	recordsCache.Set(key, records, 25*time.Minute)
	return
}

func getLatestRecord() (record Record, err error) {
	value, err := recordsCache.Get("latest")
	if err == nil {
		record = (*value)[0]
		return
	}

	err = db.Last(&record).Error
	if err != nil {
		return
	}

	record.parseConditions()
	recordsCache.Set("latest", []Record{record}, 40*time.Minute)
	return
}

func getDataPoints(measures []string, f, t *int64) (dp []DataPoint, err error) {
	for _, measure := range measures {
		if !slices.Contains(measures, measure) {
			err = errors.New("la misura richiesta non esiste")
			return
		}
	}

	if len(measures) > 5 {
		err = errors.New("sono supportate al massimo 3 misure")
		return
	}

	key := getKey(measures, f, t)

	value, err := dpCache.Get(key)
	if err == nil {
		dp = *value
		return
	}

	selectText := "dt"
	for i, measure := range measures {
		selectText += ", " + measure + " as value" + strconv.Itoa(i)
	}

	query := db.Model(&Record{})
	query = addConstraints(query.Select(selectText), f, t)

	err = query.Scan(&dp).Error
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	dpCache.Set(key, dp, 25*time.Minute)
	return
}

func initDB() (err error) {
	// Assicuriamoci che la directory "data" esista
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return errors.New("Errore nella creazione della directory 'data': " + err.Error())
	}

	// Inizializzazione del database SQLite con GORM
	db, err = gorm.Open(sqlite.Open(dbPath+dbOptions), &gorm.Config{})
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

	// Inizializzazione della zona
	zoneBytes, zErr := os.ReadFile(zonePath)
	if zErr == nil {
		zone = string(zoneBytes)
	}

	return
}
