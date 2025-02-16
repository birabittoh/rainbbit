package src

import (
	"encoding/json"
	"net/http"
)

func respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	var records []Record
	err := db.Preload("Weather").Find(&records).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, records)
}

func getServeMux() *http.ServeMux {
	s := http.NewServeMux()

	s.HandleFunc("GET /api/records", getRecords)

	return s
}
