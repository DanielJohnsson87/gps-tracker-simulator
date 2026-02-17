# Teltonika FMC003 GPS Tracker Simulator

A CLI tool that simulates a Teltonika FMC003 GPS tracker, speaking the Teltonika Codec 8 protocol over TCP. Useful for testing and developing against tracking platforms without physical hardware.

## Build

Requires Go 1.24+.

```bash
go build -o tracker .
```

## Usage

```
./tracker --server=HOST:PORT --imei=IMEI [options]
```

### Required flags

| Flag | Description |
|------|-------------|
| `--server` | Server address (host:port) |
| `--imei` | Device IMEI (15 digits) |

### Optional flags

| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `30` | Data send interval in seconds |
| `--lat` | `0.0` | Latitude |
| `--lon` | `0.0` | Longitude |
| `--altitude` | `100` | Altitude in meters |
| `--heading` | `0` | Heading in degrees (0-359) |
| `--simulation` | `stationary` | Simulation mode: stationary, buffer, pending |
| `--verbose` | `false` | Enable verbose logging |

## Simulation modes

### `stationary` (default)

Sends GPS data from a fixed position at the configured interval.

```bash
./tracker --server=localhost:5027 --imei=359633107700001 \
  --lat=54.6872 --lon=25.2797
```

### `pending`

Simulates a device that is online but has no GPS fix. Sends the given coordinates with 0 satellites, signaling an invalid position. This triggers the "Pending" state on GpsGate: the device is communicating but has no valid position (e.g. indoors or no GPS antenna).

```bash
./tracker --server=localhost:5027 --imei=359633107700001 \
  --lat=54.6872 --lon=25.2797 --simulation=pending
```

### `buffer`

Simulates a device that was offline and is flushing its buffered data. Generates 30 records with historical timestamps spaced 10 minutes apart (oldest first, starting ~290 minutes ago up to now). Records are sent one per interval, mimicking a real tracker catching up after reconnecting.

After the 30 buffered records are sent, it continues as a stationary tracker with real-time timestamps.

```bash
./tracker --server=localhost:5027 --imei=359633107700001 \
  --lat=54.6872 --lon=25.2797 --simulation=buffer --interval=5 --verbose
```

### `buffer-recentfirst`

Like `buffer`, but sends the most recent position first (current timestamp), then flushes the remaining 29 buffered records from oldest to newest. This simulates a tracker that prioritizes reporting its current location before catching up on historical data.

After all buffered records are sent, it continues as a stationary tracker with real-time timestamps.

```bash
./tracker --server=localhost:5027 --imei=359633107700001 \
  --lat=54.6872 --lon=25.2797 --simulation=buffer-recentfirst --interval=5 --verbose
```

## Protocol

Implements Teltonika Codec 8 over TCP:

1. **Connect** to server via TCP
2. **Authenticate** by sending the IMEI (2-byte length prefix + 15-byte ASCII IMEI)
3. **Send AVL data packets** containing GPS records (timestamp, coordinates, altitude, speed, heading, satellites)
4. **Receive acknowledgment** from server (4-byte record count)

Auto-reconnects on connection failure with a 5-second backoff. Graceful shutdown on SIGINT/SIGTERM (Ctrl+C).

## Project structure

```
.
├── main.go              # CLI entry point
├── protocol/
│   ├── crc.go           # CRC-16/IBM calculation
│   ├── codec8.go        # Codec 8 AVL record encoding
│   └── packet.go        # AVL packet construction
├── device/
│   └── tracker.go       # TCP connection and protocol handling
└── simulation/
    ├── generator.go     # GPSData struct and GPSGenerator interface
    ├── stationary.go    # Fixed position generator
    └── buffer.go        # Offline buffer flush simulation
```
