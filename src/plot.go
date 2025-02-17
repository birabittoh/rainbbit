package src

import (
	"bytes"
	"errors"
	"slices"
	"strings"

	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gorm.io/gorm"
)

func setAxisColor(axis *plot.Axis, color color.Color) {
	axis.Color = color
	axis.Label.TextStyle.Color = color
	axis.Tick.Color = color
	axis.Tick.Label.Color = color
}

func plotMeasure(measure string, from int64, to int64) (p *plot.Plot, err error) {
	if !slices.Contains(measures, measure) {
		err = errors.New("la misura richiesta non esiste")
		return
	}

	query := db.Model(&Record{})
	if from != 0 {
		query = query.Where("dt >= ?", from)
	}
	if to != 0 {
		query = query.Where("dt <= ?", to)
	}
	query = query.Session(&gorm.Session{})

	timestamps := make([]int64, 0)
	values := make([]float64, 0)

	err = query.Pluck("dt", &timestamps).Error
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	err = query.Pluck(measure, &values).Error
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	// Plot the data
	p = plot.New()

	p.BackgroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	setAxisColor(&p.X, color.White)
	setAxisColor(&p.Y, color.White)

	p.Title.TextStyle.Color = color.White
	p.Title.Text = strings.ReplaceAll(capitalize(measure), "_", " ")

	// Add the data to the plot (dt = x, measure = y)
	pts := make(plotter.XYs, len(timestamps))
	for i := range timestamps {
		pts[i].X = float64(timestamps[i])
		pts[i].Y = values[i]
	}

	// Add the points to the plot
	l, s, err := plotter.NewLinePoints(pts)
	if err != nil {
		err = errors.New("Errore nella creazione del plot: " + err.Error())
		return
	}

	l.Color = color.RGBA{B: 255, A: 255}
	s.Color = color.RGBA{B: 150, A: 255}
	s.GlyphStyle.Shape = draw.TriangleGlyph{}

	p.Add(l)
	p.Add(s)

	return
}

func getPlotSVG(measure string, from int64, to int64) (buf bytes.Buffer, err error) {
	p, err := plotMeasure(measure, from, to)
	if err != nil {
		err = errors.New("errore nella creazione del plot: " + err.Error())
		return
	}

	writer, err := p.WriterTo(4*vg.Inch, 4*vg.Inch, "svg")
	if err != nil {
		err = errors.New("errore nella creazione del writer SVG: " + err.Error())
		return
	}

	_, err = writer.WriteTo(&buf)
	if err != nil {
		err = errors.New("errore nella scrittura del plot SVG: " + err.Error())
		return
	}
	return
}
