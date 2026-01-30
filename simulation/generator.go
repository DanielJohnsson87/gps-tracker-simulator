package simulation

import "time"

// GPSData represents a single GPS position reading.
type GPSData struct {
	Timestamp  time.Time
	Latitude   float64
	Longitude  float64
	Altitude   int
	Speed      int // km/h
	Heading    int // degrees 0-359
	Satellites int
	Priority   byte // 0=Low, 1=High, 2=Panic
}

// GPSGenerator produces sequential GPS data points.
type GPSGenerator interface {
	Next() GPSData
}
