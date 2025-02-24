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

	bh "github.com/birabittoh/bunnyhue"
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

var (
	tmpl    map[string]*template.Template
	funcMap = template.FuncMap{
		"capitalize":      capitalize,
		"getHex":          getHex,
		"formatTimestamp": formatTimestamp,
		"getFavicon":      getFavicon,
		"getTitle":        getTitle,
	}

	palettes = map[string]*bh.Palette{
		"":      &bh.Dark, // default
		"light": &bh.Light,
	}
	themes []string
)

type PageData struct {
	Zone       string
	Palette    *bh.Palette
	FontFamily string
	OneWeekAgo int64
	Theme      string
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

func writePlot(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getLimits(r *http.Request) (from int64, to int64, palette *bh.Palette) {
	q := r.URL.Query()
	n := time.Now()

	from, err := strconv.ParseInt(q.Get("from"), 10, 64)
	if err != nil {
		from = n.Add(-24 * time.Hour).Unix()
	}
	to, err = strconv.ParseInt(q.Get("to"), 10, 64)
	if err != nil {
		to = n.Unix()
	}

	palette = getPalette(q)
	return
}

func getPalette(q url.Values) *bh.Palette {
	p, ok := palettes[q.Get("theme")]
	if ok {
		return p
	}
	return palettes[""]
}

func getPageData(q url.Values, p *bh.Palette) (*PageData, error) {
	latest, err := getLatestRecord()
	if err != nil {
		return nil, err
	}

	return &PageData{
		Zone:       zone,
		Palette:    p,
		FontFamily: fontFamily,
		OneWeekAgo: time.Now().Add(-week).Unix(),
		Theme:      q.Get("theme"),
		From:       q.Get("from"),
		To:         q.Get("to"),
		Latest:     latest,
	}, nil
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
	from, to, _ := getLimits(r)
	records, err := getAllRecords(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, records)
}

func getAPIConditions(w http.ResponseWriter, r *http.Request) {
	respond(w, conditions)
}

func getAPILatest(w http.ResponseWriter, r *http.Request) {
	latest, err := getLatestRecord()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, latest)
}

func getAPIMeta(w http.ResponseWriter, r *http.Request) {
	respond(w, map[string]any{
		"zone":     zone,
		"measures": measures,
		"themes":   themes,
	})
}

func getAPIPlot(w http.ResponseWriter, r *http.Request) {
	from, to, palette := getLimits(r)
	measure := r.PathValue("measure")
	if measure == "" {
		http.Error(w, "Misura non specificata", http.StatusBadRequest)
		return
	}

	f, t := alignConstraints(from, to)
	key := getKey([]string{measure, palette.Name}, f, t)

	value, err := plotCache.Get(key)
	if err == nil {
		writePlot(w, *value)
		return
	}

	p, err := plotMeasure(measure, f, t, palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := getPlotSVG(p, plotWidth, plotHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plotCache.Set(key, b, 25*time.Minute)
	writePlot(w, b)
}

func getAPITemp(w http.ResponseWriter, r *http.Request) {
	from, to, palette := getLimits(r)

	f, t := alignConstraints(from, to)
	key := getKey([]string{"t", palette.Name}, f, t)
	value, err := plotCache.Get(key)
	if err == nil {
		writePlot(w, *value)
		return
	}

	p, err := plotTemperature(f, t, palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := getPlotSVG(p, plotWidth, plotHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plotCache.Set(key, b, 25*time.Minute)
	writePlot(w, b)
}

func getAPIPressure(w http.ResponseWriter, r *http.Request) {
	from, to, palette := getLimits(r)

	f, t := alignConstraints(from, to)
	key := getKey([]string{"p", palette.Name}, f, t)
	value, err := plotCache.Get(key)
	if err == nil {
		writePlot(w, *value)
		return
	}

	p, err := plotPressure(f, t, palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := getPlotSVG(p, plotWidth, plotHeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	plotCache.Set(key, b, 25*time.Minute)
	writePlot(w, b)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	palette := getPalette(q)
	pd, err := getPageData(q, palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	executeTemplateSafe(w, indexPath, pd)
}

func getRecords(w http.ResponseWriter, r *http.Request) {
	from, to, palette := getLimits(r)
	records, err := getAllRecords(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pd, err := getPageData(r.URL.Query(), palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	pd.Records = records

	executeTemplateSafe(w, recordsPath, pd)
}

func getPlot(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	palette := getPalette(q)
	pd, err := getPageData(q, palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	pd.Measures = measures
	pd.Measure = r.PathValue("measure")

	executeTemplateSafe(w, plotPath, pd)
}

func parseTemplate(path string) *template.Template {
	return template.Must(template.New(path).Funcs(funcMap).ParseFiles(path, basePath))
}

func getServeMux() *http.ServeMux {
	// init templates
	tmpl = make(map[string]*template.Template)

	tmpl[indexPath] = parseTemplate(indexPath)
	tmpl[recordsPath] = parseTemplate(recordsPath)
	tmpl[plotPath] = parseTemplate(plotPath)

	for k := range palettes {
		themes = append(themes, k)
	}

	// init conditions
	var err error
	conditions, err = loadConditions()
	if err != nil {
		log.Fatalln("Errore nel caricamento delle condizioni:", err)
	}

	// init router
	s := http.NewServeMux()

	s.HandleFunc("GET /api/records", getAPIRecords)
	s.HandleFunc("GET /api/conditions", getAPIConditions)
	s.HandleFunc("GET /api/latest", getAPILatest)
	s.HandleFunc("GET /api/meta", getAPIMeta)
	s.HandleFunc("GET /api/plot/{measure}", getAPIPlot)
	s.HandleFunc("GET /api/temp", getAPITemp)
	s.HandleFunc("GET /api/pressure", getAPIPressure)

	s.HandleFunc("GET /", getIndex)
	s.HandleFunc("GET /records", getRecords)
	s.HandleFunc("GET /plot/{measure}", getPlot)
	s.HandleFunc("GET /plot/", getPlot)

	return s
}
