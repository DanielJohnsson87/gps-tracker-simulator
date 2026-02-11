package simulation

import "time"

// Pending generates GPS data that signals no valid GPS fix.
// It reports the given coordinates with 0 satellites to indicate the position
// is invalid, while keeping the device online (communicating). This triggers
// the "Pending" state on GpsGate: the device has recent activity but no recent
// valid position.
type Pending struct {
	Lat float64
	Lon float64
}

func (p *Pending) Next() GPSData {
	return GPSData{
		Timestamp:  time.Now(),
		Latitude:   p.Lat,
		Longitude:  p.Lon,
		Altitude:   0,
		Speed:      0,
		Heading:    0,
		Satellites: 0,
		Priority:   0,
	}
}
