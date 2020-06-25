package gpxcharts

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dsvg"
	"github.com/tkrajina/gpxgo/gpx"
)

type OutputExtension string

const (
	OutputSVG = ".svg"
	OutputPNG = ".png"
)

const (
	DEFAULT_LEFT_PADDING_PIXELS   = 40
	DEFAULT_BOTTOM_PADDING_PIXELS = 20
)

type Point struct {
	X, Y float64
}

type Axis struct {
	Show      bool
	Grid      float64
	Labels    float64
	FontSize  float64
	Formatter func(float64) string
}

func (a Axis) formatterOrDefault() func(float64) string {
	if a.Formatter == nil {
		return func(f float64) string {
			return fmt.Sprintf("%.2f", f)
		}
	}
	return a.Formatter
}

type Padding struct {
	Top, Right, Bottom, Left float64
}

type ChartParams struct {
	Width, Height int
	XAxis, YAxis  Axis
	Points        []Point
	FillColor     color.RGBA
	Unit          UnitType

	Padding Padding

	MinX, MaxX float64
	MinY, MaxY float64

	ChartPadding Padding

	invalid bool
}

func (cp ChartParams) UnitTypeOrMetric() UnitType {
	if cp.Unit == "" {
		return UnitTypeMetric
	}
	return cp.Unit
}

func (cp *ChartParams) calcBounds() {
	if cp.MinX == 0 && cp.MaxX == 0 {
		cp.MinX, cp.MaxX = math.MaxFloat64, -math.MaxFloat64
		for _, p := range cp.Points {
			if p.X < cp.MinX {
				cp.MinX = p.X
			}
			if p.X > cp.MaxX {
				cp.MaxX = p.X
			}
		}
	}
	if cp.MinY == 0 && cp.MaxY == 0 {
		cp.MinY, cp.MaxY = math.MaxFloat64, -math.MaxFloat64
		for _, p := range cp.Points {
			if p.Y > cp.MaxY {
				cp.MaxY = p.Y
			}
			if p.Y < cp.MinY {
				cp.MinY = p.Y
			}
		}
	}
	cp.MinX -= cp.ChartPadding.Left
	cp.MaxX += cp.ChartPadding.Right
	cp.MinY -= cp.ChartPadding.Bottom
	cp.MaxY += cp.ChartPadding.Top
}

func (cp ChartParams) toImgCoords(x, y float64) (float64, float64) {
	rx := float64(cp.Padding.Left) + float64(cp.Width-int(cp.Padding.Left)-int(cp.Padding.Right))*(x-cp.MinX)/(cp.MaxX-cp.MinX)
	ry := float64(cp.Height-int(cp.Padding.Bottom)) - float64(cp.Height-int(cp.Padding.Bottom)-int(cp.Padding.Top))*(y-cp.MinY)/(cp.MaxY-cp.MinY)
	return rx, ry
}

const (
	positive = iota
	negative
)

type ErrorLogger interface {
	Errorf(c context.Context, msg string, params ...interface{})
}

type ChartService struct {
	FontDirs       []string
	fontsFolderSet bool
	Log            ErrorLogger
}

func NewChartService(fontDirs []string) (*ChartService, error) {
	cs := ChartService{
		FontDirs: fontDirs,
	}
	if err := cs.setFontFolder(); err != nil {
		return nil, err
	}
	return &cs, nil
}

func (cs *ChartService) Errorf(c context.Context, msg string, params ...interface{}) {
	if cs.Log != nil {
		cs.Log.Errorf(c, msg, params...)
	}
}

func (cs *ChartService) setFontFolder() error {
	if cs.fontsFolderSet {
		return nil
	}
	for _, fontDir := range cs.FontDirs {
		if _, err := os.Stat(fontDir); err == nil {
			draw2d.SetFontFolder(fontDir)
			cs.fontsFolderSet = true
			return nil
		}
	}
	return errors.New("no font dir found")
}

func (cs ChartService) invalidGraphParams(c context.Context, origParams ChartParams) ChartParams {
	return ChartParams{
		invalid: true,

		Width:     origParams.Width,
		Height:    origParams.Height,
		XAxis:     Axis{Show: true},
		YAxis:     Axis{Show: true},
		FillColor: origParams.FillColor,

		Padding: origParams.Padding,

		MinX: 0,
		MinY: 0,
		MaxX: 1,
		MaxY: 1,
	}
}

func (cs ChartService) prepareSteepnesAxis(axis *Axis, max float64) {
	axis.Formatter = func(f float64) string { return fmt.Sprintf("%d°", int(math.Round(f))) }
	var g, l float64
	if max <= 5 {
		g, l = 1, 1
	} else if max <= 10 {
		g, l = 1, 2
	} else if max <= 20 {
		g, l = 5, 5
	} else {
		g, l = 5, 10
	}
	if axis.Grid == 0 {
		axis.Grid = g
	}
	if axis.Labels == 0 {
		axis.Labels = l
	}
}

func (cs ChartService) prepareSpeedAxis(axis *Axis, minSpeed, maxSpeed float64, unitType UnitType) {
	axis.Formatter = func(f float64) string { return FormatSpeed(f, unitType, true) }
	var g, l float64
	switch unitType {
	case UnitTypeNautical:
		if maxSpeed <= SPEED_KNOT*10 {
			g, l = 1*SPEED_KNOT, 2*SPEED_KNOT
		} else if maxSpeed <= SPEED_KNOT*25 {
			g, l = 2.5*SPEED_KNOT, 5*SPEED_KNOT
		} else if maxSpeed <= SPEED_KNOT*50 {
			g, l = 5*SPEED_KNOT, 10*SPEED_KNOT
		} else if maxSpeed <= SPEED_KNOT*100 {
			g, l = 10*SPEED_KNOT, 20*SPEED_KNOT
		} else {
			g, l = 25*SPEED_KNOT, 50*SPEED_KNOT
		}
	case UnitTypeImperial:
		if maxSpeed <= SPEED_MPH*10 {
			g, l = 1*SPEED_MPH, 2*SPEED_MPH
		} else if maxSpeed <= SPEED_MPH*25 {
			g, l = 2.5*SPEED_MPH, 5*SPEED_MPH
		} else if maxSpeed <= SPEED_MPH*50 {
			g, l = 5*SPEED_MPH, 10*SPEED_MPH
		} else if maxSpeed <= SPEED_MPH*100 {
			g, l = 10*SPEED_MPH, 20*SPEED_MPH
		} else {
			g, l = 25*SPEED_MPH, 50*SPEED_MPH
		}
	default:
		if maxSpeed <= 10*SPEED_KMH {
			g, l = 1*SPEED_KMH, 2*SPEED_KMH
		} else if maxSpeed <= 25*SPEED_KMH {
			g, l = 2.5*SPEED_KMH, 5*SPEED_KMH
		} else if maxSpeed <= 50*SPEED_KMH {
			g, l = 5*SPEED_KMH, 10*SPEED_KMH
		} else if maxSpeed <= 100*SPEED_KMH {
			g, l = 10*SPEED_KMH, 20*SPEED_KMH
		} else {
			g, l = 25*SPEED_KMH, 50*SPEED_KMH
		}
	}
	if axis.Grid == 0 {
		axis.Grid = g
	}
	if axis.Labels == 0 {
		axis.Labels = l
	}
}

func (cs ChartService) prepareElevationAxis(axis *Axis, minEle, maxEle float64, unitType UnitType) {
	length := maxEle - minEle

	var g, l float64
	axis.Formatter = func(f float64) string { return FormatAltitude(f, unitType) }
	switch unitType {
	case UnitTypeImperial, UnitTypeNautical:
		if length <= ONE_FEET*50 {
			g, l = 10*ONE_FEET, 10*ONE_FEET
		} else if length <= ONE_FEET*100 {
			g, l = 20*ONE_FEET, 20*ONE_FEET
		} else if length <= ONE_FEET*500 {
			g, l = 100*ONE_FEET, 50*ONE_FEET
		} else if length <= ONE_FEET*1000 {
			g, l = 200*ONE_FEET, 100*ONE_FEET
		} else if length <= ONE_FEET*4000 {
			g, l = 500*ONE_FEET, 500*ONE_FEET
		} else if length <= ONE_FEET*10000 {
			g, l = 1000*ONE_FEET, 1000*ONE_FEET
		} else if length <= ONE_FEET*100000 {
			g, l = 10000*ONE_FEET, 10000*ONE_FEET
		} else {
			g, l = 10000*ONE_FEET, 10000*ONE_FEET
		}
	default:
		if length <= 100 {
			g, l = 20, 20
		} else if length <= 500 {
			g, l = 50, 100
		} else if length <= 1000 {
			g, l = 100, 200
		} else if length <= 2000 {
			g, l = 125, 250
		} else {
			g, l = 250, 500
		}
	}
	if axis.Grid == 0 {
		axis.Grid = g
	}
	if axis.Labels == 0 {
		axis.Labels = l
	}
}

func (cs ChartService) prepareLengthAxis(axis *Axis, length float64, unitType UnitType) {
	axis.Formatter = func(f float64) string { return FormatLength(f, unitType) }
	var g, l float64
	switch unitType {
	case UnitTypeImperial:
		if length <= ONE_MILE {
			g, l = 0.1*ONE_MILE, 0.1*ONE_MILE
		} else if length <= ONE_MILE*4 {
			g, l = 0.5*ONE_MILE, 0.5*ONE_MILE
		} else if length <= ONE_MILE*10 {
			g, l = 1*ONE_MILE, 0.5*ONE_MILE
		} else if length <= ONE_MILE*20 {
			g, l = 2*ONE_MILE, 1*ONE_MILE
		} else if length <= ONE_MILE*50 {
			g, l = 5*ONE_MILE, 2.5*ONE_MILE
		} else if length <= ONE_MILE*100 {
			g, l = 10*ONE_MILE, 5*ONE_MILE
		} else if length <= ONE_MILE*200 {
			g, l = 10*ONE_MILE, 5*ONE_MILE
		} else if length <= ONE_MILE*400 {
			g, l = 25*ONE_MILE, 5*ONE_MILE
		} else {
			g, l = 100*ONE_MILE, 100*ONE_MILE
		}
	case UnitTypeNautical:
		if length <= ONE_NAUTICAL_MILE {
			g, l = 0.1*ONE_NAUTICAL_MILE, 0.1*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*4 {
			g, l = 0.5*ONE_NAUTICAL_MILE, 0.5*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*10 {
			g, l = 1*ONE_NAUTICAL_MILE, 0.5*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*25 {
			g, l = 2*ONE_NAUTICAL_MILE, 1*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*50 {
			g, l = 5*ONE_NAUTICAL_MILE, 2.5*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*100 {
			g, l = 10*ONE_NAUTICAL_MILE, 5*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*200 {
			g, l = 10*ONE_NAUTICAL_MILE, 5*ONE_NAUTICAL_MILE
		} else if length <= ONE_NAUTICAL_MILE*400 {
			g, l = 25*ONE_NAUTICAL_MILE, 5*ONE_NAUTICAL_MILE
		} else {
			g, l = 100*ONE_NAUTICAL_MILE, 100*ONE_NAUTICAL_MILE
		}
	default: //case UnitTypeMetric:
		if length <= 1000 {
			g, l = 100, 100
		} else if length <= 4000 {
			g, l = 500, 500
		} else if length <= 10000 {
			g, l = 500, 1000
		} else if length <= 20000 {
			g, l = 1000, 2000
		} else if length <= 50000 {
			g, l = 1000, 2000
		} else if length <= 100000 {
			g, l = 2500, 5000
		} else if length <= 200000 {
			g, l = 5000, 10000
		} else if length <= 400000 {
			g, l = 5000, 25000
		} else {
			g, l = 50000, 100000
		}
	}
	if axis.Grid == 0 {
		axis.Grid = g
	}
	if axis.Labels == 0 {
		axis.Labels = l
	}
}

func (cs ChartService) chart(c context.Context, params ChartParams, output OutputExtension) ([]byte, error) {
	var gc draw2d.GraphicContext
	switch output {
	case OutputPNG:
		img := image.NewRGBA(image.Rect(0, 0, params.Width, params.Height))
		gc := draw2dimg.NewGraphicContext(img)
		cs.renderChart(c, params, gc)
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("error encoding png %w", err)
		}
		return buf.Bytes(), nil
	case OutputSVG:
		svg := draw2dsvg.NewSvg()
		gc = draw2dsvg.NewGraphicContext(svg)
		cs.renderChart(c, params, gc)
		bytes, err := xml.Marshal(svg)
		if err != nil {
			return nil, fmt.Errorf("error marshalling svg %w", err)
		}
		return bytes, nil
	default:
		return nil, fmt.Errorf("invalid format %s", output)
	}
}

func (cs ChartService) renderChart(c context.Context, params ChartParams, gc draw2d.GraphicContext) {
	// Initialize the graphic context on an RGBA image
	//rect := image.Rect(0, 0, params.Width, params.Height)
	//dest := image.NewRGBA(rect)

	// Background:
	gc.BeginPath()
	gc.MoveTo(0, 0)
	gc.SetStrokeColor(color.RGBA{0xFF, 0xFF, 0xFF, 0xff})
	gc.SetLineWidth(0)
	gc.LineTo(float64(params.Width), 0.0)
	gc.LineTo(float64(params.Width), float64(params.Height))
	gc.LineTo(0, float64(params.Height))
	gc.Close()
	gc.FillStroke()

	params.calcBounds()
	//fmt.Printf("params=%#v\n", params)

	if params.MinX >= params.MaxX {
		cs.Errorf(c, "minx=%f, maxX=%f", params.MinX, params.MaxX)
		params = cs.invalidGraphParams(c, params)
	}
	if params.MinY >= params.MaxY {
		cs.Errorf(c, "minY=%f, maxY=%f", params.MinY, params.MaxY)
		params = cs.invalidGraphParams(c, params)
	}
	if IsNanOrOnf(params.MinX) || IsNanOrOnf(params.MaxX) {
		cs.Errorf(c, "minx=%f, maxX=%f", params.MinX, params.MaxX)
		params = cs.invalidGraphParams(c, params)
	}
	if IsNanOrOnf(params.MinY) || IsNanOrOnf(params.MaxY) {
		cs.Errorf(c, "minY=%f, maxY=%f", params.MinY, params.MaxY)
		params = cs.invalidGraphParams(c, params)
	}

	// Grid:
	if params.XAxis.Grid > 0 {
		grid := params.XAxis.Grid
		for v := grid * float64(int(params.MinX/grid)); v < params.MaxX; v += grid {
			if v < params.MinX {
				continue
			}
			gc.BeginPath()
			gc.MoveTo(params.toImgCoords(v, params.MinY))
			gc.SetStrokeColor(color.RGBA{0xE0, 0xE0, 0xE0, 0xff})
			gc.SetLineWidth(0.5)
			gc.LineTo(params.toImgCoords(v, params.MaxY))
			gc.Close()
			gc.FillStroke()
		}
	}
	if params.YAxis.Grid > 0 {
		grid := params.YAxis.Grid
		for v := grid * float64(int(params.MinY/grid)); v < params.MaxY; v += grid {
			if v < params.MinY {
				continue
			}
			gc.BeginPath()
			gc.MoveTo(params.toImgCoords(params.MinX, v))
			gc.SetStrokeColor(color.RGBA{0xE0, 0xE0, 0xE0, 0xff})
			gc.SetLineWidth(0.5)
			gc.LineTo(params.toImgCoords(params.MaxX, v))
			gc.Close()
			gc.FillStroke()
		}
	}

	// Graph:
	for _, pn := range []int{positive, negative} {
		for n, point := range params.Points {
			x, y := point.X, point.Y
			if n == 0 {
				gc.BeginPath() // Initialize a new path
				gc.MoveTo(params.toImgCoords(x, params.MinY))
				gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xaf})
				gc.SetFillColor(color.RGBA{0x10, 0x10, 0x10, 0x40})
				gc.SetLineWidth(0.5)
			}

			switch pn {
			case positive:
				if y < 0 {
					y = 0
				}
			case negative:
				if y > 0 {
					y = 0
				}
			}

			if y < params.MinY {
				y = params.MinY
			}

			gc.LineTo(params.toImgCoords(x, y))

			if n == len(params.Points)-1 {
				gc.LineTo(params.toImgCoords(x, math.Max(0.0, params.MinY)))
				gc.LineTo(params.toImgCoords(0, math.Max(0.0, params.MinY)))
				gc.Close()
				gc.FillStroke()
			}
		}
	}

	fontSize := 8.0
	fontData := draw2d.FontData{Name: "luxi", Family: draw2d.FontFamilySerif, Style: draw2d.FontStyleBold | draw2d.FontStyleItalic}

	axisColor := color.RGBA{0x36, 0x6a, 0xff, 0xff}
	if params.XAxis.Show {
		if params.XAxis.FontSize > 0 {
			fontSize = params.XAxis.FontSize
		}
		gc.BeginPath()
		gc.MoveTo(params.toImgCoords(params.MinX, params.MinY))
		gc.SetStrokeColor(axisColor)
		gc.SetLineWidth(0.5)
		gc.LineTo(params.toImgCoords(params.MaxX, params.MinY))
		gc.Close()
		gc.FillStroke()
		labels := params.XAxis.Labels
		if labels > 0 {
			for v := labels * float64(int(params.MinX/labels)); v < params.MaxX; v += labels {
				if v == 0 {
					continue
				}
				gc.BeginPath()
				x, y := params.toImgCoords(v, params.MinY)
				gc.MoveTo(x, y-3)
				gc.SetStrokeColor(axisColor)
				gc.SetLineWidth(0.5)
				gc.LineTo(x, y+3)
				gc.Close()
				gc.FillStroke()

				gc.SetFontData(fontData)

				txt := params.XAxis.formatterOrDefault()(v)

				gc.SetFontData(fontData)
				gc.SetFillColor(color.RGBA{0, 0, 0, 0})
				textWidth := gc.FillStringAt(txt, x-fontSize, float64(y)+float64(fontSize+4))

				gc.SetFillColor(axisColor)
				gc.SetFontSize(fontSize)
				gc.FillStringAt(txt, x-textWidth/2, float64(y)+float64(fontSize+4))
			}
		}
	}
	if params.YAxis.Show {
		if params.YAxis.FontSize > 0 {
			fontSize = params.YAxis.FontSize
		}
		gc.BeginPath()
		gc.MoveTo(params.toImgCoords(params.MinX, params.MinY))
		gc.SetStrokeColor(axisColor)
		gc.SetLineWidth(0.5)
		gc.LineTo(params.toImgCoords(params.MinX, params.MaxY))
		gc.Close()
		gc.FillStroke()
		labels := params.YAxis.Labels
		if labels > 0 {
			for v := labels * float64(int(params.MinY/labels)); v < params.MaxY; v += labels {
				if v == 0 {
					continue
				}
				if v < params.MinY {
					continue
				}
				gc.BeginPath()
				x, y := params.toImgCoords(params.MinX, v)
				gc.MoveTo(x-3, y)
				gc.SetStrokeColor(axisColor)
				gc.SetLineWidth(0.5)
				gc.LineTo(x+3, y)
				gc.Close()
				gc.FillStroke()

				gc.SetFontSize(fontSize)

				txt := params.YAxis.formatterOrDefault()(v)
				gc.SetFontData(fontData)
				gc.SetFillColor(color.RGBA{0, 0, 0, 0})
				textWidth := gc.FillStringAt(txt, x-fontSize, float64(y)+float64(fontSize+4))

				gc.SetFillColor(axisColor)
				gc.FillStringAt(txt, x-textWidth-fontSize/2, y+fontSize/2)
			}
		}
	}

	if params.invalid {
		x, y := params.toImgCoords((params.MinX+params.MaxX)/2, (params.MinY+params.MaxY)/2)
		txt := "No enough data available"
		gc.SetFontData(fontData)
		gc.SetFillColor(color.RGBA{0, 0, 0, 0})
		textWidth := gc.FillStringAt(txt, x-fontSize, float64(y)+float64(fontSize+4))
		gc.SetFillColor(color.RGBA{0xff, 0x4e, 0x00, 0xff})
		gc.FillStringAt(txt, x-textWidth/2, float64(y)+float64(fontSize+4))
	}
}

func (cs ChartService) SpeedChart(c context.Context, params ChartParams, g gpx.GPX, output OutputExtension) ([]byte, error) {
	g.ReduceTrackPoints(1000, 50)
	var points []Point
	var minSpeed, maxSpeed float64
	var d float64
	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			for n, pt := range segment.Points {
				if 0 < n {
					d += pt.Distance2D(&segment.Points[n-1])
				}
				if 0 < n && n < len(segment.Points)-1 {
					prevPt := segment.Points[n-1]
					pt := segment.Points[n]
					nextPt := segment.Points[n+1]
					if !prevPt.Timestamp.IsZero() && !pt.Timestamp.IsZero() && !nextPt.Timestamp.IsZero() {
						duration := nextPt.Timestamp.Sub(prevPt.Timestamp)
						length := nextPt.Distance2D(&pt) + pt.Distance2D(&prevPt)
						speed := length / duration.Seconds()
						if len(points) == 0 || speed < minSpeed {
							minSpeed = speed
						}
						if len(points) == 0 || speed > maxSpeed {
							maxSpeed = speed
						}
						points = append(points, Point{d, speed})
					}
				}
			}
		}
	}
	params.Points = points
	cs.prepareLengthAxis(&params.XAxis, d, params.UnitTypeOrMetric())
	cs.prepareSpeedAxis(&params.YAxis, minSpeed, maxSpeed, params.UnitTypeOrMetric())
	return cs.chart(c, params, output)
}

func (cs ChartService) SteepnessChart(c context.Context, params ChartParams, g gpx.GPX, output OutputExtension) ([]byte, error) {
	g.ReduceTrackPoints(1000, 50)
	g.SmoothVertical()
	g.SmoothVertical()
	g.SmoothVertical()
	g.SmoothVertical()
	var points []Point
	var d float64
	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			for n, pt := range segment.Points {
				if n > 0 {
					d += pt.Distance2D(&segment.Points[n-1])
				}
				angle := 0.0
				if 0 < n && n < len(segment.Points)-1 {
					prevPt := segment.Points[n-1]
					nextPt := segment.Points[n+1]
					if prevPt.Elevation.NotNull() && nextPt.Elevation.NotNull() {
						angle = gpx.ElevationAngle(prevPt.Point, nextPt.Point, false)
					}
				}
				points = append(points, Point{d, angle})
			}
		}
	}

	sumFrom0 := 0.0
	for _, pt := range points {
		sumFrom0 += math.Abs(pt.Y)
	}
	max := 4 * sumFrom0 / float64(len(points))

	params.MinY, params.MaxY = -max, max
	params.Points = points
	cs.prepareLengthAxis(&params.XAxis, d, params.UnitTypeOrMetric())
	cs.prepareSteepnesAxis(&params.YAxis, max)
	return cs.chart(c, params, output)
}

func (cs ChartService) ElevationChart(c context.Context, params ChartParams, g gpx.GPX, output OutputExtension) ([]byte, error) {
	var (
		minElevation = 1000.0
		maxElevation = 0.0
		points       []Point
		d            float64
	)
	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			for n, pt := range segment.Points {
				if n > 0 {
					d += pt.Distance2D(&segment.Points[n-1])
				}
				ele := pt.Elevation.Value()
				points = append(points, Point{d, ele})

				if ele < minElevation {
					minElevation = ele
				}
				if ele > maxElevation {
					maxElevation = ele
				}
			}
		}
	}

	params.Points = points
	cs.prepareLengthAxis(&params.XAxis, d, params.UnitTypeOrMetric())
	cs.prepareElevationAxis(&params.YAxis, minElevation, maxElevation, params.UnitTypeOrMetric())
	return cs.chart(c, params, output)
}
