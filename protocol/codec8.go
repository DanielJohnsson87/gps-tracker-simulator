package protocol

import (
	"encoding/binary"

	"tracker/simulation"
)

// AVLRecord is a fully encoded Codec 8 AVL record ready for packet assembly.
type AVLRecord struct {
	Data []byte
}

// EncodeAVLRecord encodes a GPSData point into a Codec 8 AVL record.
func EncodeAVLRecord(gps simulation.GPSData) AVLRecord {
	var buf []byte
	buf = append(buf, buildTimestamp(gps)...)
	buf = append(buf, gps.Priority)
	buf = append(buf, buildGPSElement(gps)...)
	buf = append(buf, buildIOElement()...)
	return AVLRecord{Data: buf}
}

// buildTimestamp encodes the timestamp as 8 bytes (milliseconds since Unix epoch).
func buildTimestamp(gps simulation.GPSData) []byte {
	ms := uint64(gps.Timestamp.UnixMilli())
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ms)
	return b
}

// buildGPSElement encodes the 15-byte GPS element.
//
//	Longitude:  4 bytes (degrees * 10,000,000, signed via two's complement)
//	Latitude:   4 bytes (degrees * 10,000,000, signed via two's complement)
//	Altitude:   2 bytes (meters, unsigned)
//	Angle:      2 bytes (degrees, unsigned)
//	Satellites: 1 byte
//	Speed:      2 bytes (km/h, unsigned)
func buildGPSElement(gps simulation.GPSData) []byte {
	b := make([]byte, 15)

	lon := int32(gps.Longitude * 10_000_000)
	lat := int32(gps.Latitude * 10_000_000)

	binary.BigEndian.PutUint32(b[0:4], uint32(lon))
	binary.BigEndian.PutUint32(b[4:8], uint32(lat))
	binary.BigEndian.PutUint16(b[8:10], uint16(gps.Altitude))
	binary.BigEndian.PutUint16(b[10:12], uint16(gps.Heading))
	b[12] = byte(gps.Satellites)
	binary.BigEndian.PutUint16(b[13:15], uint16(gps.Speed))

	return b
}

// buildIOElement builds an empty I/O element (no I/O properties).
// Codec 8 I/O element format:
//
//	Event IO ID:     1 byte  (0x00 = no event)
//	Total IO count:  1 byte  (0x00)
//	1-byte IO count: 1 byte  (0x00)
//	2-byte IO count: 1 byte  (0x00)
//	4-byte IO count: 1 byte  (0x00)
//	8-byte IO count: 1 byte  (0x00)
func buildIOElement() []byte {
	return []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}
