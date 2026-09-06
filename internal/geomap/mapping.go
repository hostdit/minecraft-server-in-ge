package geomap

import "math"

const metresPerDegLat = 111320.0

type Anchor struct {
	Lat, Lng        float64
	BlockX, BlockZ  int
	MetresPerBlock  float64
}

func (a Anchor) ToGeo(x, z int) (lat, lng float64) {
	metresEast := float64(x-a.BlockX) * a.MetresPerBlock
	metresNorth := -float64(z-a.BlockZ) * a.MetresPerBlock
	lat = a.Lat + metresNorth/metresPerDegLat
	lng = a.Lng + metresEast/(metresPerDegLat*math.Cos(a.Lat*math.Pi/180))
	return lat, lng
}

func (a Anchor) ToBlock(lat, lng float64) (x, z int) {
	metresNorth := (lat - a.Lat) * metresPerDegLat
	metresEast := (lng - a.Lng) * metresPerDegLat * math.Cos(a.Lat*math.Pi/180)
	x = a.BlockX + int(math.Round(metresEast/a.MetresPerBlock))
	z = a.BlockZ - int(math.Round(metresNorth/a.MetresPerBlock))
	return x, z
}

func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	p := math.Pi / 180
	dLat := (lat2 - lat1) * p
	dLng := (lng2 - lng1) * p
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*p)*math.Cos(lat2*p)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Sqrt(a))
}
