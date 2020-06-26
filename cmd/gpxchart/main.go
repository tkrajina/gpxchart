package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tkrajina/go-elevations/geoelevations"
	"github.com/tkrajina/gpxchart/gpxcharts"
	"github.com/tkrajina/gpxgo/gpx"
)

const OptsBackupExtension = ".gpxchars_opts"

type GraphType string

const (
	Elevation GraphType = "elevation"
	Speed     GraphType = "speed"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	if len(os.Args) == 2 && strings.HasSuffix(os.Args[1], OptsBackupExtension) {
		var newArgs []string
		byts, err := ioutil.ReadFile(os.Args[1])
		panicIfErr(err)
		panicIfErr(json.Unmarshal(byts, &newArgs))
		fmt.Printf("Using options: %s\n", strings.Join(newArgs, " "))
		os.Args = append([]string{os.Args[0]}, newArgs...)
	}

	c := context.Background()
	var (
		params           gpxcharts.ChartParams
		typ              string
		outputFile       string
		fontSize         string
		help             bool
		size             string
		padding          string
		grid             string
		labels           string
		chartPadding     string
		imperial         bool
		debug            bool
		srtm             bool
		smoothElevations bool
	)

	flag.BoolVar(&help, "help", false, "Help")
	flag.StringVar(&size, "s", "900,200", "Size (width,height)")
	flag.StringVar(&grid, "g", "0,0", "Grid lines (x,y)")
	flag.StringVar(&labels, "l", "0,0", "Labels (x,y)")
	flag.StringVar(&padding, "p", "40,20,0,0", "Padding (left,down,right,up)")
	flag.StringVar(&chartPadding, "cp", "20,5,20,10", "Chart padding (left,down,right,up)")
	flag.StringVar(&fontSize, "f", "8,8", "Both axes font size (x,y)")
	flag.StringVar(&typ, "t", string(Elevation), fmt.Sprintf("Type (%s or %s)", Elevation, Speed))
	flag.StringVar(&outputFile, "o", "chart.svg", "Output filename (.png or .svg)")
	flag.BoolVar(&imperial, "im", false, "Use imperial units (mi, ft)")
	flag.BoolVar(&srtm, "srtm", false, "Overwrite elevations from SRTM")
	flag.BoolVar(&smoothElevations, "sme", false, "Smooth elevations")
	flag.BoolVar(&imperial, "d", false, "Debug")
	flag.Parse()

	if help {
		showHelp()
		return
	}

	if imperial {
		params.Unit = gpxcharts.UnitTypeImperial
	}
	params.Width, params.Height = twoInts(size)
	params.XAxis.FontSize, params.YAxis.FontSize = twoFloats(fontSize)
	params.XAxis.Grid, params.YAxis.Grid = twoFloats(grid)
	params.XAxis.Labels, params.YAxis.Labels = twoFloats(labels)
	params.XAxis.Show = true
	params.YAxis.Show = true
	params.Padding = gpxcharts.Padding{
		Top: 0, Right: 0, Bottom: 20, Left: 40,
	}

	cs, err := gpxcharts.NewChartService([]string{prepareFont()})
	if err != nil {
		panic(err)
	}

	params.Padding.Left, params.Padding.Bottom, params.Padding.Right, params.Padding.Top = fourFloats(padding)
	params.ChartPadding.Left, params.ChartPadding.Bottom, params.ChartPadding.Right, params.ChartPadding.Top = fourFloats(chartPadding)

	if debug {
		fmt.Printf("params=%#v\n", params)
	}

	var chartGen func(c context.Context, params gpxcharts.ChartParams, g gpx.GPX, output gpxcharts.OutputExtension) ([]byte, error)
	switch GraphType(typ) {
	case Elevation:
		chartGen = cs.ElevationChart
	case Speed:
		chartGen = cs.SpeedChart
	default:
		showHelp()
		os.Exit(1)
	}

	if len(flag.Args()) != 1 {
		fmt.Printf("Expected one file, found: %d\n", len(flag.Args()))
		showHelp()
		os.Exit(1)
	}

	file := flag.Args()[0]
	g, err := gpx.ParseFile(file)
	if err != nil {
		panic("Error loading: " + file)
	}

	if srtm {
		overwriteElevations(g)
	}
	if smoothElevations {
		for i := 0; i < 4; i++ {
			g.SmoothVertical()
		}
	}

	bytes, err := chartGen(c, params, *g, gpxcharts.OutputExtension(filepath.Ext(outputFile)))
	panicIfErr(err)
	err = ioutil.WriteFile(outputFile, bytes, 0700)
	panicIfErr(err)

	byts, err := json.MarshalIndent(os.Args[1:], "", "    ")
	panicIfErr(err)
	ioutil.WriteFile(file+OptsBackupExtension, byts, 0700)
	fmt.Printf("Saved opions file %s\n", file+OptsBackupExtension)
	fmt.Printf("Saved chart to %s\n", outputFile)
}

func overwriteElevations(g *gpx.GPX) error {
	srtm, err := geoelevations.NewSrtm(http.DefaultClient)
	if err != nil {
		return err
	}

	g.ExecuteOnAllPoints(func(pt *gpx.GPXPoint) {
		ele, err := srtm.GetElevation(http.DefaultClient, pt.Latitude, pt.Longitude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot get elevation for %f,%f: %v", pt.Latitude, pt.Longitude, err)
		}
		fmt.Printf("Overwriting elevation for %f,%f: %f\n", pt.Latitude, pt.Longitude, ele)
		pt.Elevation = *gpx.NewNullableFloat64(ele)
	})

	for i := 0; i < 5; i++ {
		g.SmoothVertical()
	}

	return nil
}

func showHelp() {
	fmt.Println()
	flag.Usage()
	fmt.Println()
}

func prepareFont() string {
	home, err := os.UserHomeDir()
	panicIfErr(err)
	dir := path.Join(home, ".cache", "gpxcharts", "font")
	panicIfErr(os.MkdirAll(dir, 0700))
	file := path.Join(dir, "luxisr.ttf")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("Saving font file to", file)
		panicIfErr(ioutil.WriteFile(file, LuxiFont, 0700))
	}
	return dir
}

func twoInts(str string) (int, int) {
	f1, f2 := twoFloats(str)
	return int(f1), int(f2)
}

func twoFloats(str string) (float64, float64) {
	floats := parseFloats(str)
	if len(floats) != 2 {
		panic(fmt.Sprintf("Invalid 2 numbers: %s", str))
	}
	return floats[0], floats[1]
}

func fourFloats(str string) (float64, float64, float64, float64) {
	floats := parseFloats(str)
	if len(floats) != 4 {
		panic(fmt.Sprintf("Invalid 4 numbers: %s", str))
	}
	return floats[0], floats[1], floats[2], floats[3]
}

func parseFloats(str string) []float64 {
	var res []float64
	for n, part := range strings.Split(str, ",") {
		f1, err := strconv.ParseFloat(part, 32)
		if err != nil {
			panic(fmt.Sprintf("Invalid number %d in %s", n+1, part))
		}
		res = append(res, f1)
	}
	return res
}
