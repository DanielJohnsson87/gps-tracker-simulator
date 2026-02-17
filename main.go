package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tracker/device"
	"tracker/simulation"
)

func main() {
	server := flag.String("server", "", "Server address (host:port)")
	imei := flag.String("imei", "", "Device IMEI (15 digits)")
	interval := flag.Int("interval", 30, "Data send interval in seconds")
	lat := flag.Float64("lat", 0.0, "Starting latitude")
	lon := flag.Float64("lon", 0.0, "Starting longitude")
	altitude := flag.Int("altitude", 100, "Altitude in meters")
	heading := flag.Int("heading", 0, "Initial heading in degrees (0-359)")
	sim := flag.String("simulation", "", "Simulation mode: buffer, pending, buffer-recentfirst, stationary (default: stationary)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Teltonika FMC003 GPS Tracker Simulator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --server=HOST:PORT --imei=IMEI [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --server=localhost:5027 --imei=359633107700001\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --server=192.168.1.100:5027 --imei=359633107700001 --lat=40.7128 --lon=-74.0060 --verbose\n", os.Args[0])
	}

	flag.Parse()

	if *server == "" {
		fmt.Fprintln(os.Stderr, "Error: --server is required")
		flag.Usage()
		os.Exit(1)
	}
	if *imei == "" || len(*imei) != 15 {
		fmt.Fprintln(os.Stderr, "Error: --imei is required and must be 15 digits")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Teltonika Tracker Simulator")
	log.Printf("  Server:   %s", *server)
	log.Printf("  IMEI:     %s", *imei)
	log.Printf("  Interval: %ds", *interval)
	log.Printf("  Position: %.6f, %.6f", *lat, *lon)
	log.Printf("  Altitude: %dm", *altitude)
	log.Printf("  Heading:  %d°", *heading)
	log.Printf("  Simulation: %s", *sim)
	log.Printf("  Verbose:  %v", *verbose)

	var gen simulation.GPSGenerator
	switch *sim {
	case "buffer":
		log.Printf("  Mode: buffer (30 records, 10 min apart, oldest first)")
		gen = simulation.NewBuffer(*lat, *lon, *altitude, *heading)
	case "pending":
		gen = &simulation.Pending{Lat: *lat, Lon: *lon}
	case "buffer-recentfirst":
		log.Printf("  Mode: buffer-recentfirst (most recent first, then 29 oldest-to-newest)")
		gen = simulation.NewRecentFirst(*lat, *lon, *altitude, *heading)
	case "stationary", "":
		gen = &simulation.Stationary{
			Lat:      *lat,
			Lon:      *lon,
			Altitude: *altitude,
			Heading:  *heading,
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown simulation mode %q\n", *sim)
		flag.Usage()
		os.Exit(1)
	}

	tracker := &device.Tracker{
		Server:  *server,
		IMEI:    *imei,
		Verbose: *verbose,
	}

	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Shutting down...")
		close(stop)
	}()

	tracker.Run(gen, time.Duration(*interval)*time.Second, stop)
	tracker.Close()
	log.Println("Stopped.")
}
