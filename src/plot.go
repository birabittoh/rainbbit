package src

import (
	"errors"
	"slices"
	"strings"

	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gorm.io/gorm"
)

type PlotData struct {
	Timestamps []int64
	Values     []float64
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
