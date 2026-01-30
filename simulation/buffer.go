package simulation

import "time"

const (
	BufferSize     = 30
	BufferInterval = 10 * time.Minute
)

// Buffer is a GPSGenerator that simulates a device flushing its offline buffer.
// It produces 30 records with historical timestamps (oldest first), then
// switches to real-time timestamps for subsequent calls.
type Buffer struct {
	Lat      float64
	Lon      float64
	Altitude int
	Heading  int

	queue []time.Time
	index int
}

// NewBuffer creates a Buffer generator pre-loaded with 30 historical timestamps.
// The oldest record is 290 minutes ago, each subsequent record is 10 minutes
// closer to now, and the 30th record has the current time.
func NewBuffer(lat, lon float64, altitude, heading int) *Buffer {
	now := time.Now()
	queue := make([]time.Time, BufferSize)
	for i := 0; i < BufferSize; i++ {
		offset := time.Duration(BufferSize-1-i) * BufferInterval
		queue[i] = now.Add(-offset)
	}
	return &Buffer{
		Lat:      lat,
		Lon:      lon,
		Altitude: altitude,
		Heading:  heading,
		queue:    queue,
	}
}

func (b *Buffer) Next() GPSData {
	var ts time.Time
	if b.index < len(b.queue) {
		ts = b.queue[b.index]
		b.index++
	} else {
		ts = time.Now()
	}

	return GPSData{
		Timestamp:  ts,
		Latitude:   b.Lat,
		Longitude:  b.Lon,
		Altitude:   b.Altitude,
		Speed:      0,
		Heading:    b.Heading,
		Satellites: 12,
		Priority:   0,
	}
}
