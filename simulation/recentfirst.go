package simulation

import "time"

// RecentFirst simulates a device flushing its offline buffer but sending the
// most recent position first, followed by buffered records from oldest to newest.
// After all buffered records are sent, it continues with real-time timestamps.
type RecentFirst struct {
	Lat      float64
	Lon      float64
	Altitude int
	Heading  int

	queue []time.Time
	index int
}

// NewRecentFirst creates a RecentFirst generator pre-loaded with 30 historical
// timestamps. The first call to Next returns the most recent timestamp (now),
// then subsequent calls return the remaining 29 records from oldest to newest.
func NewRecentFirst(lat, lon float64, altitude, heading int) *RecentFirst {
	now := time.Now()
	// Build queue: first element is now (most recent), then oldest-to-newest.
	queue := make([]time.Time, BufferSize)
	queue[0] = now
	for i := 1; i < BufferSize; i++ {
		offset := time.Duration(BufferSize-1-i) * BufferInterval
		queue[i] = now.Add(-offset)
	}
	return &RecentFirst{
		Lat:      lat,
		Lon:      lon,
		Altitude: altitude,
		Heading:  heading,
		queue:    queue,
	}
}

func (r *RecentFirst) Next() GPSData {
	var ts time.Time
	if r.index < len(r.queue) {
		ts = r.queue[r.index]
		r.index++
	} else {
		ts = time.Now()
	}

	return GPSData{
		Timestamp:  ts,
		Latitude:   r.Lat,
		Longitude:  r.Lon,
		Altitude:   r.Altitude,
		Speed:      0,
		Heading:    r.Heading,
		Satellites: 12,
		Priority:   0,
	}
}
