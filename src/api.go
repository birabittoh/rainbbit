package src

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	base = "base"
	ps   = string(os.PathSeparator)

	basePath    = "templates" + ps + base + ".gohtml"
	indexPath   = "templates" + ps + "index.gohtml"
	recordsPath = "templates" + ps + "records.gohtml"
	plotPath    = "templates" + ps + "plot.gohtml"

	week = 24 * 7 * time.Hour
)

var tmpl map[string]*template.Template

type PageData struct {
	FontFamily string
	OneWeekAgo int64
	From       string
	To         string
	Measure    string
	Measures   []string
	Records    []Record
	Latest     Record
}

func respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func getLimits(r *http.Request) (q url.Values, from int64, to int64) {
	q = r.URL.Query()

	from, err := strconv.ParseInt(q.Get("from"), 10, 64)
	if err != nil {
		from = time.Now().Add(-24 * time.Hour).Unix()
	}
	to, err = strconv.ParseInt(q.Get("to"), 10, 64)
	if err != nil {
		to = 0
	}

	return
}

func getPageData(q url.Values) *PageData {
	return &PageData{
		FontFamily: fontFamily,
		OneWeekAgo: time.Now().Add(-week).Unix(),
		From:       q.Get("from"),
		To:         q.Get("to"),
	}
}

func executeTemplateSafe(w http.ResponseWriter, t string, data any) {
	var buf bytes.Buffer
	if err := tmpl[t].ExecuteTemplate(&buf, base, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func getAPIRecords(w http.ResponseWriter, r *http.Request) {
	_, from, to := getLimits(r)
	records, err := getAllRecords(from, to)
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
	_, from, to := getLimits(r)
	measure := r.PathValue("measure")
	if measure == "" {
		http.Error(w, "Misura non specificata", http.StatusBadRequest)
		return
	}

	p, err := plotMeasure(measure, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf, err := getPlotSVG(p, plotWidth, plotHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func getAPITemp(w http.ResponseWriter, r *http.Request) {
	_, from, to := getLimits(r)

	p, err := plotTemperature(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf, err := getPlotSVG(p, plotWidth, plotHeight)
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

	pd := getPageData(r.URL.Query())
	pd.Latest = latest

	executeTemplateSafe(w, indexPath, pd)
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	q, from, to := getLimits(r)
	records, err := getAllRecords(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pd := getPageData(q)
	pd.Records = records

	executeTemplateSafe(w, recordsPath, pd)
}

func getPlot(w http.ResponseWriter, r *http.Request) {
	pd := getPageData(r.URL.Query())
	pd.Measures = measures
	pd.Measure = r.PathValue("measure")

	executeTemplateSafe(w, plotPath, pd)
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
	s.HandleFunc("GET /api/temp", getAPITemp)

	s.HandleFunc("GET /", getIndex)
	s.HandleFunc("GET /records", getRecords)
	s.HandleFunc("GET /plot/{measure}", getPlot)
	s.HandleFunc("GET /plot/", getPlot)

	return s
}
