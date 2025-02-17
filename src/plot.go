package src

import (
	"bytes"
	"errors"

	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type DataPoint struct {
	Dt     float64
	Value0 float64
	Value1 float64
	Value2 float64
	Value3 float64
	Value4 float64
}

func setAxisColor(axis *plot.Axis, color color.Color) {
	axis.Color = color
	axis.Label.TextStyle.Color = color
	axis.Tick.Color = color
	axis.Tick.Label.Color = color
}

func plotMeasure(measure string, from int64, to int64) (p *plot.Plot, err error) {
	dp, err := getDataPoints([]string{measure}, from, to)
	if err != nil {
		err = errors.New("errore nella lettura dei dati: " + err.Error())
		return
	}

	// Plot the data
	p = plot.New()

	p.BackgroundColor = color.Transparent
	setAxisColor(&p.X, color.White)
	setAxisColor(&p.Y, color.White)

	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04:05"}

	p.Title.TextStyle.Color = color.White
	p.Title.Text = measure

	pts := make(plotter.XYs, len(dp))
	for i := range dp {
		pts[i].X = dp[i].Dt
		pts[i].Y = dp[i].Value0
	}

	// Add the points to the plot
	l, s, err := plotter.NewLinePoints(pts)
	if err != nil {
		err = errors.New("Errore nella creazione del plot: " + err.Error())
		return
	}

	l.Color = color.Gray{Y: 128}
	s.Color = color.Gray{Y: 128}
	s.GlyphStyle.Shape = draw.CircleGlyph{}

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
