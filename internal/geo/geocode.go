package geo

import (
	_ "embed"
	"encoding/json"
	"math"
	"sync"

	"gemc/internal/geomap"
)

//go:embed countries.min.json
var countriesRaw []byte

//go:embed cities.min.json
var citiesRaw []byte

type country struct {
	N   string        `json:"n"`
	ISO string        `json:"iso"`
	P   [][][][2]float64 `json:"p"`
}

type city struct {
	Name    string
	Country string
	Lat     float64
	Lng     float64
}

var (
	once      sync.Once
	countries []country
	cities    []city
)

func load() {
	once.Do(func() {
		_ = json.Unmarshal(countriesRaw, &countries)
		var raw [][]any
		_ = json.Unmarshal(citiesRaw, &raw)
		cities = make([]city, 0, len(raw))
		for _, r := range raw {
			if len(r) < 4 {
				continue
			}
			name, _ := r[0].(string)
			cn, _ := r[1].(string)
			lat, _ := r[2].(float64)
			lng, _ := r[3].(float64)
			cities = append(cities, city{name, cn, lat, lng})
		}
	})
}

func pointInRing(lng, lat float64, ring [][2]float64) bool {
	in := false
	n := len(ring)
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := ring[i][0], ring[i][1]
		xj, yj := ring[j][0], ring[j][1]
		if (yi > lat) != (yj > lat) &&
			lng < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			in = !in
		}
		j = i
	}
	return in
}

func Country(lat, lng float64) string {
	load()
	for _, c := range countries {
		for _, poly := range c.P {
			if len(poly) == 0 {
				continue
			}
			if pointInRing(lng, lat, poly[0]) {
				hole := false
				for _, ring := range poly[1:] {
					if pointInRing(lng, lat, ring) {
						hole = true
						break
					}
				}
				if !hole {
					return c.N
				}
			}
		}
	}
	return ""
}

func NearestCity(lat, lng float64) (string, float64) {
	load()
	best := ""
	bestD := math.Inf(1)
	for _, c := range cities {
		d := geomap.HaversineKm(lat, lng, c.Lat, c.Lng)
		if d < bestD {
			bestD = d
			best = c.Name
		}
	}
	return best, bestD
}

func Describe(lat, lng float64) string {
	country := Country(lat, lng)
	cityName, d := NearestCity(lat, lng)
	switch {
	case country == "" && cityName == "":
		return "the open ocean"
	case country == "":
		return "the sea near " + cityName
	case d < 25:
		return cityName + ", " + country
	default:
		return country
	}
}
