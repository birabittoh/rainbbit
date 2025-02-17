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

func getPlotSVG(p *plot.Plot) (buf bytes.Buffer, err error) {
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

func addLines(p *plot.Plot, points plotter.XYs, color color.Color, dashed bool, label string) error {
	l, s, err := plotter.NewLinePoints(points)
	if err != nil {
		return errors.New("Errore nella creazione del plot: " + err.Error())
	}

	if dashed {
		l.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	}
	l.Color = color
	s.Color = color
	s.GlyphStyle.Shape = draw.CircleGlyph{}

	p.Add(l)
	p.Add(s)

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
	p = plot.New()

	p.BackgroundColor = color.Transparent
	setAxisColor(&p.X, color.White)
	setAxisColor(&p.Y, color.White)

	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04:05"}

	p.Title.Text = "Temperature"
	p.Title.TextStyle.Color = color.White
	p.Legend.TextStyle.Color = color.White

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
	}

	tColor := color.Gray{Y: 128}
	tMinColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	tMaxColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	flColor := color.Gray{Y: 50}

	// Add the plot points to the plot
	err = addLines(p, tPts, tColor, false, "Temp")
	if err != nil {
		return
	}
	err = addLines(p, tMinPts, tMinColor, false, "Min")
	if err != nil {
		return
	}
	err = addLines(p, tMaxPts, tMaxColor, false, "Max")
	if err != nil {
		return
	}
	err = addLines(p, flPts, flColor, true, "Feels Like")
	if err != nil {
		return
	}

	return
}
