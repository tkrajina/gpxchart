# GPX Charts

A command-line tool and library for elevation charts from GPX files.

## Installation

Install Golang and:

    go install github.com/tkrajina/gpxchart/cmd/gpxchart/...

## Usage

```
Usage of gpxchart:
  -cp string
        Chart padding (left,down,right,up) (default "20,5,20,10")
  -d    Debug
  -f string
        Both axes font size (x,y) (default "8,8")
  -g string
        Grid lines (x,y) (default "0,0")
  -help
        Help
  -im
        Use imperial units (mi, ft)
  -l string
        Labels (x,y) (default "0,0")
  -o string
        Output filename (.png or .svg) (default "chart.svg")
  -p string
        Padding (left,down,right,up) (default "40,20,0,0")
  -s string
        Size (width,height) (default "900,200")
  -t string
        Type (elevation or speed) (default "elevation")
```

## Examples



### Simple

`gpxchart -o examples/simple.png  test_files/zbevnica.gpx`

![Simple](examples/simple.png)

### With smoothed elevations

`gpxchart -o examples/smoothed.png -sme test_files/zbevnica.gpx`

![With smoothed elevations](examples/smoothed.png)

### With SRTM elevations

`gpxchart -o examples/with_srtm_elevations.png -srtm test_files/zbevnica.gpx`

![With SRTM elevations](examples/with_srtm_elevations.png)

### SVG output

`gpxchart -o examples/simple.svg -s 200,100 test_files/zbevnica.gpx`

![SVG output](examples/simple.svg)

### Imperial units

`gpxchart -o examples/imperial.png -im test_files/zbevnica.gpx`

![Imperial units](examples/imperial.png)

### Custom size

`gpxchart -o examples/custom_size.png -s 900,300 test_files/zbevnica.gpx`

![Custom size](examples/custom_size.png)

### No padding

`gpxchart -o examples/no_padding.png -p 0,0,0,0 test_files/zbevnica.gpx`

![No padding](examples/no_padding.png)

### Padding

`gpxchart -o examples/custom_padding.png -p 50,20,20,20 test_files/zbevnica.gpx`

![Padding](examples/custom_padding.png)

### Custom font size

`gpxchart -o examples/custom_font_size.png -p 100,20,0,0 -f 10,20 test_files/zbevnica.gpx`

![Custom font size](examples/custom_font_size.png)

### Custom grid

`gpxchart -o examples/custom_grid.png -g 50,20 test_files/zbevnica.gpx`

![Custom grid](examples/custom_grid.png)

### Custom labels

`gpxchart -o examples/custom_labels.png -l 250,20 test_files/zbevnica.gpx`

![Custom labels](examples/custom_labels.png)

### Custom chart padding

`gpxchart -o examples/custom_chart_padding.png -cp 500,50,500,50 test_files/zbevnica.gpx`

![Custom chart padding](examples/custom_chart_padding.png)






# License

**gpxcharts** is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)
