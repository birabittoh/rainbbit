package src

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	base = "base"
	ps   = string(os.PathSeparator)

	basePath    = "templates" + ps + base + ".html"
	indexPath   = "templates" + ps + "index.html"
	recordsPath = "templates" + ps + "records.html"
	plotPath    = "templates" + ps + "plot.html"
)

var tmpl map[string]*template.Template

type PlotData struct {
	From     string
	To       string
	Measure  string
	Measures []string
}

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

func getAPIMeasures(w http.ResponseWriter, r *http.Request) {
	respond(w, measures)
}

func getAPIPlot(w http.ResponseWriter, r *http.Request) {
	measure := r.PathValue("measure")
	if measure == "" {
		http.Error(w, "Misura non specificata", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	from, err := strconv.ParseInt(q.Get("from"), 10, 64)
	if err != nil {
		from = 0
	}
	to, err := strconv.ParseInt(q.Get("to"), 10, 64)
	if err != nil {
		to = 0
	}

	buf, err := getPlotSVG(measure, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	latest, err := getLatestRecord()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl[indexPath].ExecuteTemplate(w, base, latest)
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	records, err := getAllRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl[recordsPath].ExecuteTemplate(w, base, records)
}

func getPlot(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pd := PlotData{
		From:     q.Get("from"),
		To:       q.Get("to"),
		Measure:  r.PathValue("measure"),
		Measures: measures,
	}

	tmpl[plotPath].ExecuteTemplate(w, base, pd)
}

func getServeMux() *http.ServeMux {
	// init templates
	tmpl = make(map[string]*template.Template)

	tmpl[indexPath] = template.Must(template.ParseFiles(indexPath, basePath))
	tmpl[recordsPath] = template.Must(template.ParseFiles(recordsPath, basePath))
	tmpl[plotPath] = template.Must(template.ParseFiles(plotPath, basePath))

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
	s.HandleFunc("GET /api/measures", getAPIMeasures)
	s.HandleFunc("GET /api/plot/{measure}", getAPIPlot)

	s.HandleFunc("GET /", getIndex)
	s.HandleFunc("GET /records", getRecords)
	s.HandleFunc("GET /plot/{measure}", getPlot)

	return s
}
