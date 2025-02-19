package src

import (
	"bytes"
	"errors"
	"math"
	"time"

	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	plotWidth  = 7 * vg.Inch
	plotHeight = 4 * vg.Inch
	fontFamily = "Arial, sans-serif"
	tickFormat = "15:04 02/01"
)

var (
	dc = map[string]color.Color{
		"lightGray":    color.RGBA{R: 224, G: 224, B: 224, A: 255},
		"dodgerBlue":   color.RGBA{R: 30, G: 144, B: 255, A: 255},
		"redOrange":    color.RGBA{R: 255, G: 69, B: 0, A: 255},
		"limeGreen":    color.RGBA{R: 50, G: 205, B: 50, A: 255},
		"gold":         color.RGBA{R: 255, G: 215, B: 0, A: 255},
		"orchid":       color.RGBA{R: 218, G: 112, B: 214, A: 255},
		"mediumPurple": color.RGBA{R: 147, G: 112, B: 219, A: 255},
		"cyan":         color.RGBA{R: 0, G: 255, B: 255, A: 255},
	}

	plotFont = font.Font{
		Typeface: "Liberation",
		Variant:  "Sans",
		Size:     10,
	}
)

type DataPoint struct {
	Dt     float64
	Value0 float64
	Value1 float64
	Value2 float64
	Value3 float64
	Value4 float64
}

// customTimeTicks implementa plot.Ticker
type customTimeTicks struct {
	times []time.Time
}

func (ct customTimeTicks) Ticks(min, max float64) []plot.Tick {
	// Prima raccogliamo tutti i timestamp validi (quelli compresi in [min, max])
	var validTimes []time.Time
	for _, t := range ct.times {
		unix := float64(t.Unix())
		if unix < min || unix > max {
			continue
		}
		validTimes = append(validTimes, t)
	}

	// Impostiamo il numero massimo di tick che vogliamo mostrare.
	const maxTicks = 10
	var ticks []plot.Tick

	// Se il numero di timestamp validi è minore o uguale a maxTicks, li stampiamo tutti.
	if len(validTimes) <= maxTicks {
		for _, t := range validTimes {
			ticks = append(ticks, plot.Tick{
				Value: float64(t.Unix()),
				Label: t.Format(tickFormat),
			})
		}
		return ticks
	}

	// Se ci sono più timestamp, calcoliamo un intervallo per "saltare" alcuni tick.
	step := len(validTimes) / maxTicks
	if step < 1 {
		step = 1
	}

	// Selezioniamo ogni 'step'-esimo timestamp.
	for i, t := range validTimes {
		if i%step == 0 {
			ticks = append(ticks, plot.Tick{
				Value: float64(t.Unix()),
				Label: t.Format(tickFormat),
			})
		}
	}

	// Assicuriamoci di includere anche l'ultimo timestamp, se non è già stato aggiunto.
	last := validTimes[len(validTimes)-1]
	if len(ticks) == 0 || ticks[len(ticks)-1].Value != float64(last.Unix()) {
		ticks = append(ticks, plot.Tick{
			Value: float64(last.Unix()),
			Label: last.Format(tickFormat),
		})
	}

	return ticks
}

func setAxisColor(axis *plot.Axis, color color.Color) {
	axis.Color = color
	axis.Label.TextStyle.Color = color
	axis.Tick.Color = color
	axis.Tick.Label.Color = color
}

func newDarkPlot(timestamps []time.Time) *plot.Plot {
	p := plot.New()
	p.BackgroundColor = color.Transparent
	setAxisColor(&p.X, color.White)
	setAxisColor(&p.Y, color.White)
	p.X.Tick.Marker = customTimeTicks{times: timestamps}
	p.X.Tick.Label.Rotation = math.Pi / -2
	p.X.Tick.Label.XAlign = 0.05
	p.X.Tick.Label.YAlign = 0
	p.Title.TextStyle.Color = color.White
	p.Legend.TextStyle.Color = color.White
	p.Legend.TextStyle.Font = plotFont
	p.X.Tick.Label.Font = plotFont
	p.Y.Tick.Label.Font = plotFont
	return p
}

func getPlotSVG(p *plot.Plot, w vg.Length, h vg.Length) (buf bytes.Buffer, err error) {
	writer, err := p.WriterTo(w, h, "svg")
	if err != nil {
		err = errors.New("errore nella creazione del writer SVG: " + err.Error())
		return
	}

	_, err = writer.WriteTo(&buf)
	if err != nil {
		err = errors.New("errore nella scrittura del plot SVG: " + err.Error())
		return
	}

	buf = *bytes.NewBuffer(bytes.ReplaceAll(buf.Bytes(), []byte("Liberation Sans"), []byte(fontFamily)))
	return
}

func plotMeasure(measure string, from int64, to int64) (p *plot.Plot, err error) {
	dp, err := getDataPoints([]string{measure}, from, to)
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	var timestamps []time.Time
	pts := make(plotter.XYs, len(dp))
	for i := range dp {
		pts[i].X = dp[i].Dt
		pts[i].Y = dp[i].Value0

		timestamps = append(timestamps, time.Unix(int64(dp[i].Dt), 0).Round(time.Minute))
	}

	// Plot the data
	p = newDarkPlot(timestamps)

	addLines(p, pts, dc["lightGray"], false, capitalize(measure))

	return
}

func addLines(p *plot.Plot, points plotter.XYs, color color.Color, dashed bool, label string) error {
	l, err := plotter.NewLine(points)
	if err != nil {
		return errors.New("Errore nella creazione del plot: " + err.Error())
	}

	if dashed {
		l.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	}
	l.Color = color
	p.Add(l)

	p.Legend.Add(label, l)
	return nil
}

func plotTemperature(from int64, to int64) (p *plot.Plot, err error) {
	dp, err := getDataPoints([]string{"temp", "temp_min", "temp_max", "feels_like"}, from, to)
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	// Plot the data
	var timestamps []time.Time

	tPts := make(plotter.XYs, len(dp))
	tMinPts := make(plotter.XYs, len(dp))
	tMaxPts := make(plotter.XYs, len(dp))
	flPts := make(plotter.XYs, len(dp))
	for i := range dp {
		tPts[i].X = dp[i].Dt
		tMinPts[i].X = dp[i].Dt
		tMaxPts[i].X = dp[i].Dt
		flPts[i].X = dp[i].Dt

		tPts[i].Y = dp[i].Value0
		tMinPts[i].Y = dp[i].Value1
		tMaxPts[i].Y = dp[i].Value2
		flPts[i].Y = dp[i].Value3

		timestamps = append(timestamps, time.Unix(int64(dp[i].Dt), 0).Round(time.Minute))
	}

	p = newDarkPlot(timestamps)

	// Add the plot points to the plot
	err = addLines(p, flPts, dc["gold"], true, "Feels Like")
	if err != nil {
		return
	}
	err = addLines(p, tPts, dc["lightGray"], false, "Temp")
	if err != nil {
		return
	}
	err = addLines(p, tMinPts, dc["dodgerBlue"], false, "Min")
	if err != nil {
		return
	}
	err = addLines(p, tMaxPts, dc["redOrange"], false, "Max")
	if err != nil {
		return
	}

	return
}
