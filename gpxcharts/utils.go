package gpxcharts

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/png"
	"math"
	"strings"

	"github.com/llgcode/draw2d/draw2dsvg"
)

func SVGToBytes(svg *draw2dsvg.Svg) ([]byte, error) {
	byts, err := xml.Marshal(svg)
	if err != nil {
		return nil, errors.New("error marhslling")
	}
	return byts, nil
}

func RGBAToBytes(m *image.RGBA) ([]byte, error) {
	var b bytes.Buffer
	if err := png.Encode(&b, m); err != nil {
		return nil, errors.New("error encoding png")
	}
	return b.Bytes(), nil
}

func FormatSpeed(meters_per_seconds float64, unit_type UnitType, round bool) string {
	if meters_per_seconds <= 0 {
		return "n/a"
	}
	if len(unit_type) == 0 {
		unit_type = UnitTypeMetric
	}

	var (
		speed float64
		unit  string
	)
	if unit_type == UnitTypeImperial {
		speed = meters_per_seconds * 60 * 60 / UnitTypeImperial.Units()["mi"]
		unit = "mph"
	} else if unit_type == UnitTypeNautical {
		speed = meters_per_seconds * 60 * 60 / 1852.
		unit = "kn"
	} else {
		speed = meters_per_seconds * 60 * 60 / 1000.
		unit = "kmh"
	}

	if round {
		return fmt.Sprintf("%d%s", int(math.Round(speed)), unit)
	}

	if speed < 10 {
		return fmt.Sprintf("%.2f%s", speed, unit)
	}
	return fmt.Sprintf("%.1f%s", speed, unit)
}

func FormatLength(lengthM float64, ut UnitType) string {
	if lengthM < 0 {
		return "n/a"
	}

	if len(ut) == 0 {
		ut = UnitTypeMetric
	}

	if ut == UnitTypeNautical {
		miles := ConvertFromM(lengthM, "NM")
		if miles < 10 {
			return FormatFloat(miles, 2) + "NM"
		} else {
			return FormatFloat(miles, 1) + "NM"
		}
	} else if ut == UnitTypeImperial {
		miles := ConvertFromM(lengthM, "mi")
		if miles < 10 {
			return FormatFloat(miles, 2) + "mi"
		} else {
			return FormatFloat(miles, 1) + "mi"
		}
	} else { // metric:
		if lengthM < 1000 {
			return FormatFloat(lengthM, 0) + "m"
		}
		if lengthM < 50000 {
			return FormatFloat(lengthM/1000, 2) + "km"
		}
	}

	return FormatFloat(lengthM/1000, 1) + "km"
}

// Convert from meters (or m/s if speed) into...
func ConvertFromM(n float64, toUnit string) float64 {
	toUnit = strings.TrimSpace(strings.ToLower(toUnit))
	if v, is := SPEED_UNITS[toUnit]; is {
		return n / v
	}
	if v, is := Units[toUnit]; is {
		return n / v
	}
	return 0
}

func FormatFloat(f float64, digits int) string {
	format := fmt.Sprintf("%%.%df", digits)
	res := fmt.Sprintf(format, f)
	if strings.Contains(res, ".") {
		res = strings.TrimRight(res, "0")
	}
	return strings.TrimRight(res, ".")
}

func FormatAltitude(altitude_m float64, unit_type UnitType) string {
	if altitude_m < -20000 || altitude_m > 20000 {
		return "n/a"
	}
	if unit_type == UnitTypeMetric {
		return FormatFloat(altitude_m, 0) + "m"
	}
	return FormatFloat(ConvertFromM(altitude_m, "ft"), 0) + "ft"
}

func IsNanOrOnf(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0)
}
