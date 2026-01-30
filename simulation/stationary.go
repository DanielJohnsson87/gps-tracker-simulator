package simulation

import "time"

// Stationary generates GPS data for a fixed position with minor random drift.
type Stationary struct {
	Lat      float64
	Lon      float64
	Altitude int
	Heading  int
}

func (s *Stationary) Next() GPSData {
	return GPSData{
		Timestamp:  time.Now(),
		Latitude:   s.Lat,
		Longitude:  s.Lon,
		Altitude:   s.Altitude,
		Speed:      0,
		Heading:    s.Heading,
		Satellites: 12,
		Priority:   0,
	}
}
