package src

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func templatePath(s string) string {
	return filepath.Join("templates", s) + ".html"
}

var tmpl map[string]*template.Template

func respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func getAPIRecords(w http.ResponseWriter, r *http.Request) {
	records, err := getAllRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, records)
}

func getAPILatest(w http.ResponseWriter, r *http.Request) {
	latest, err := getLatestRecord()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, latest)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	latest, err := getLatestRecord()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl[templatePath("index")].ExecuteTemplate(w, "base", latest)
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	records, err := getAllRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl[templatePath("records")].ExecuteTemplate(w, "base", records)
}

func getServeMux() *http.ServeMux {
	// init templates
	tmpl = make(map[string]*template.Template)
	tmpl[templatePath("index")] = template.Must(template.ParseFiles(templatePath("index"), templatePath("base")))
	tmpl[templatePath("records")] = template.Must(template.ParseFiles(templatePath("records"), templatePath("base")))

	// init conditions
	var err error
	conditions, err = loadConditions()
	if err != nil {
		log.Fatalln("Errore nel caricamento delle condizioni:", err)
	}

	// init router
	s := http.NewServeMux()

	s.HandleFunc("GET /api/records", getAPIRecords)
	s.HandleFunc("GET /api/latest", getAPILatest)

	s.HandleFunc("GET /", getIndex)
	s.HandleFunc("GET /records", getRecords)

	return s
}
