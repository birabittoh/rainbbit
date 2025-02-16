package src

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
	Rain1H float64 `json:"rain_1h"`
	Snow1H float64 `json:"snow_1h"`

	// Relazione 1 a N con WeatherRecord
	Weather []Weather `json:"-" gorm:"foreignKey:RecordDt"`

	Conditions []Condition `json:"conditions" gorm:"-"`
	Favicon    string      `json:"-" gorm:"-"`
	Title      string      `json:"-" gorm:"-"`
	TimeAgo    string      `json:"-" gorm:"-"`
}

// Weather rappresenta un elemento dell'array weather
type Weather struct {
	ID        uint `json:"id" gorm:"primarykey"`
	RecordDt  uint `json:"record_dt"`
	WeatherID int  `json:"weather_id"`
}

func getAllRecords() (records []Record, err error) {
	err = db.Preload("Weather").Find(&records).Error
	for i := range records {
		records[i].parseConditions()
	}
	return
}

func getLatestRecord() (record Record, err error) {
	err = db.Preload("Weather").Last(&record).Error
	record.parseConditions()
	return
}
