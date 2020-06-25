package gpxcharts

import (
	"math"
	"strings"
)

const (
	ONE_FEET          = 0.3048
	ONE_YARD          = 0.9144
	ONE_MILE          = 1.609344 * 1000
	ONE_NAUTICAL_MILE = 1852
)

type UnitType string

func (ut UnitType) Name() string {
	switch ut {
	case UnitTypeMetric:
		return "metric"
	case UnitTypeImperial:
		return "imperial"
	case UnitTypeNautical:
		return "nautical"
	default:
		return "unknown"
	}
}

func (ut UnitType) Units() map[string]float64 {
	switch ut {
	case UnitTypeImperial:
		return map[string]float64{
			"ft": ONE_FEET,
			"yd": ONE_YARD,
			"mi": ONE_MILE,
		}
	case UnitTypeNautical:
		return map[string]float64{
			"NM": ONE_NAUTICAL_MILE,
		}
	default:
		return map[string]float64{
			"m":  1.,
			"km": 1000.,
			"cm": 0.01,
		}
	}
}

const (
	UnitTypeMetric   UnitType = "m"
	UnitTypeImperial UnitType = "i"
	UnitTypeNautical UnitType = "n"
)

func AllUnitTypes() []UnitType {
	return []UnitType{
		UnitTypeMetric,
		UnitTypeImperial,
		UnitTypeNautical,
	}
}

var (
	SPEED_MPS  = 1.
	SPEED_KMH  = 1000. / math.Pow(60., 2)
	SPEED_MPH  = 1.609344 * 1000. / math.Pow(60., 2)
	SPEED_KNOT = 1852. / math.Pow(60., 2)
)

var (
	Units       = map[string]float64{}
	SPEED_UNITS = map[string]float64{
		"mps":  SPEED_MPS,
		"kmh":  SPEED_KMH,
		"mph":  SPEED_MPH,
		"knot": SPEED_KNOT,
	}
)

func init() {
	for _, ut := range AllUnitTypes() {
		for k, v := range ut.Units() {
			Units[strings.ToLower(k)] = v
		}
	}
}
