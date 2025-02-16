package src

import (
	"log"
	"time"
)

// ------------------------
// MODELLI GORM
// ------------------------

// Record rappresenta i dati meteo completi (tranne la slice Weather, salvata separatamente)
type Record struct {
	Dt         time.Time `json:"dt" gorm:"primarykey"`
	Visibility int       `json:"visibility"`

	// Sys
	Sunrise time.Time `json:"sunrise"`
	Sunset  time.Time `json:"sunset"`

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
	Rain1H float64 `json:"rain_1h"`
	Snow1H float64 `json:"snow_1h"`

	// Relazione 1 a N con WeatherRecord
	Weather []Weather `json:"-" gorm:"foreignKey:RecordDt"`

	Conditions []Condition `json:"conditions" gorm:"-"`
	Favicon    string      `json:"-" gorm:"-"`
	Title      string      `json:"-" gorm:"-"`
}

// Weather rappresenta un elemento dell'array weather
type Weather struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	RecordDt  time.Time `json:"record_dt"`
	WeatherID int       `json:"weather_id"`
}

func getAllRecords() (records []Record, err error) {
	err = db.Preload("Weather").Find(&records).Error
	return
}

func getLatestRecord() (record Record, err error) {
	err = db.Preload("Weather").Last(&record).Error
	for _, w := range record.Weather {
		c, ok := conditions[w.WeatherID]
		if !ok {
			log.Printf("Condizione meteo non trovata per ID %d\n", w.WeatherID)
			continue
		}

		if record.Dt.After(record.Sunrise) && record.Dt.Before(record.Sunset) {
			c.Icon += "d"
		} else {
			c.Icon += "n"
		}

		c.Description = capitalize(c.Description)

		record.Conditions = append(record.Conditions, c)
	}

	record.Favicon = record.Conditions[0].Icon
	record.Title = record.Conditions[0].Description
	return
}
