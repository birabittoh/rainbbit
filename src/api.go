package src

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"gonum.org/v1/plot/vg"
)

const (
	base = "base"

	basePath    = "templates" + string(os.PathSeparator) + base + ".html"
	indexPath   = "templates" + string(os.PathSeparator) + "index.html"
	recordsPath = "templates" + string(os.PathSeparator) + "records.html"
)

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

	p, err := plotMeasure(measure, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the plot to an SVG file
	writer, err := p.WriterTo(4*vg.Inch, 4*vg.Inch, "svg")
	if err != nil {
		http.Error(w, "Errore nella creazione del writer SVG", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	_, err = writer.WriteTo(&buf)
	if err != nil {
		http.Error(w, "Errore nella scrittura del plot SVG", http.StatusInternalServerError)
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

func getServeMux() *http.ServeMux {
	// init templates
	tmpl = make(map[string]*template.Template)

	tmpl[indexPath] = template.Must(template.ParseFiles(indexPath, basePath))
	tmpl[recordsPath] = template.Must(template.ParseFiles(recordsPath, basePath))

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

	return s
}
