package device

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"tracker/protocol"
	"tracker/simulation"
)

// Tracker manages the TCP connection to a Teltonika-compatible server
// and sends periodic AVL data packets.
type Tracker struct {
	Server  string
	IMEI    string
	Verbose bool

	conn net.Conn
}

// Connect establishes a TCP connection to the server.
func (t *Tracker) Connect() error {
	t.logf("Connecting to %s...", t.Server)
	conn, err := net.DialTimeout("tcp", t.Server, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	t.conn = conn
	t.logf("Connected to %s", t.Server)
	return nil
}

// SendIMEI sends the IMEI authentication message and reads the server response.
// The IMEI message is: 2 bytes length (0x000F) + 15 bytes IMEI ASCII.
// Server responds with 1 byte: 0x01 = accepted, 0x00 = rejected.
func (t *Tracker) SendIMEI() error {
	// Build IMEI packet: 2-byte length prefix + IMEI string
	imeiBytes := []byte(t.IMEI)
	msg := make([]byte, 2, 2+len(imeiBytes))
	binary.BigEndian.PutUint16(msg, uint16(len(imeiBytes)))
	msg = append(msg, imeiBytes...)

	t.logf("Sending IMEI: %s", t.IMEI)
	if _, err := t.conn.Write(msg); err != nil {
		return fmt.Errorf("send IMEI: %w", err)
	}

	// Read server response
	t.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	resp := make([]byte, 1)
	if _, err := t.conn.Read(resp); err != nil {
		return fmt.Errorf("read IMEI response: %w", err)
	}
	t.conn.SetReadDeadline(time.Time{})

	if resp[0] != 0x01 {
		return fmt.Errorf("IMEI rejected by server (response: 0x%02X)", resp[0])
	}

	t.logf("IMEI accepted by server")
	return nil
}

// SendAVLData sends an AVL data packet and reads the server acknowledgment.
// Server responds with 4 bytes: the number of accepted records (big-endian).
func (t *Tracker) SendAVLData(packet []byte) (int, error) {
	if _, err := t.conn.Write(packet); err != nil {
		return 0, fmt.Errorf("send AVL data: %w", err)
	}

	// Read acknowledgment (4 bytes = number of accepted records)
	t.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	resp := make([]byte, 4)
	if _, err := t.conn.Read(resp); err != nil {
		return 0, fmt.Errorf("read AVL response: %w", err)
	}
	t.conn.SetReadDeadline(time.Time{})

	accepted := int(binary.BigEndian.Uint32(resp))
	return accepted, nil
}

// Close closes the TCP connection.
func (t *Tracker) Close() {
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
}

// Run starts the main loop: connects to the server, authenticates with IMEI,
// and sends GPS data at the specified interval. Reconnects automatically on failure.
// The stop channel is used for graceful shutdown.
func (t *Tracker) Run(gen simulation.GPSGenerator, interval time.Duration, stop <-chan struct{}) {
	for {
		if err := t.connectAndAuthenticate(); err != nil {
			log.Printf("Error: %v", err)
			if !t.waitOrStop(5*time.Second, stop) {
				return
			}
			continue
		}

		t.sendLoop(gen, interval, stop)
		t.Close()

		// Check if we should stop or reconnect
		select {
		case <-stop:
			return
		default:
			log.Println("Connection lost, reconnecting...")
			if !t.waitOrStop(5*time.Second, stop) {
				return
			}
		}
	}
}

func (t *Tracker) connectAndAuthenticate() error {
	if err := t.Connect(); err != nil {
		return err
	}
	if err := t.SendIMEI(); err != nil {
		t.Close()
		return err
	}
	return nil
}

func (t *Tracker) sendLoop(gen simulation.GPSGenerator, interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send first data point immediately
	if !t.sendOne(gen) {
		return
	}

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if !t.sendOne(gen) {
				return
			}
		}
	}
}

// sendOne generates a GPS point, encodes it, and sends it. Returns false on error.
func (t *Tracker) sendOne(gen simulation.GPSGenerator) bool {
	gps := gen.Next()
	record := protocol.EncodeAVLRecord(gps)
	packet := protocol.BuildAVLPacket([]protocol.AVLRecord{record})

	t.logf("Sending AVL data: lat=%.6f lon=%.6f alt=%d speed=%d heading=%d",
		gps.Latitude, gps.Longitude, gps.Altitude, gps.Speed, gps.Heading)

	accepted, err := t.SendAVLData(packet)
	if err != nil {
		log.Printf("Error sending AVL data: %v", err)
		return false
	}

	t.logf("Server accepted %d record(s)", accepted)
	return true
}

// waitOrStop waits for the given duration or until stop is signaled.
// Returns true if the wait completed, false if stop was signaled.
func (t *Tracker) waitOrStop(d time.Duration, stop <-chan struct{}) bool {
	select {
	case <-stop:
		return false
	case <-time.After(d):
		return true
	}
}

func (t *Tracker) logf(format string, args ...any) {
	if t.Verbose {
		log.Printf(format, args...)
	}
}
