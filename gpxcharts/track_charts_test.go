package gpxcharts

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrajina/gpxgo/gpx"
)

var chartService *ChartService

func init() {
	cs, err := NewChartService([]string{"../fonts"})
	if err != nil {
		panic(err)
	}
	chartService = cs
}

func TestChart(t *testing.T) {
	t.Parallel()

	gpxFiles := []string{
		"../test_files/track.gpx",
		"../test_files/parenzana.gpx",
		"../test_files/empty.gpx",
	}
	chartFuncs := []func(c context.Context, params ChartParams, g gpx.GPX, output OutputExtension) ([]byte, error){
		chartService.ElevationChart,
		chartService.SpeedChart,
		chartService.SteepnessChart,
	}
	unitTypes := []UnitType{
		UnitTypeMetric,
		UnitTypeImperial,
		UnitTypeNautical,
	}

	for n, fn := range gpxFiles {
		for m, f := range chartFuncs {
			for _, unit := range unitTypes {
				for _, output := range []OutputExtension{OutputPNG, OutputSVG} {
					//fmt.Printf("%s and func #%d %#v\n", fn, n, f)
					bytes := testChart(t, fn, f, unit, output)
					if t.Failed() {
						t.FailNow()
					}
					f, err := os.Create(fmt.Sprintf("tmp_chart_%s_%d_%d%s", unit, n, m, output))
					assert.Nil(t, err)
					_, err = f.Write(bytes)
					assert.Nil(t, err)
					f.Close()
				}
			}
		}
	}
}

func testChart(t *testing.T, fileName string, chartFunc func(c context.Context, params ChartParams, g gpx.GPX, output OutputExtension) ([]byte, error), unit UnitType, output OutputExtension) []byte {
	g, err := gpx.ParseFile(fileName)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	params := ChartParams{
		Padding: Padding{Left: DEFAULT_LEFT_PADDING_PIXELS, Bottom: DEFAULT_BOTTOM_PADDING_PIXELS},
		XAxis: Axis{
			Show:   true,
			Grid:   500,
			Labels: 1000,
			Formatter: func(f float64) string {
				return FormatAltitude(f, UnitTypeMetric)
			},
		},
		YAxis: Axis{
			Show:   true,
			Grid:   25,
			Labels: 50,
			Formatter: func(f float64) string {
				return FormatLength(f, UnitTypeMetric)
			},
		},
		Width:  1000,
		Height: 250,
		Unit:   unit,
	}

	byts, err := chartFunc(context.Background(), params, *g, output)
	assert.Nil(t, err)

	return byts
}
