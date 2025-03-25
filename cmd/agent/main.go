package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Read configuration
	config.SetDefaults()
	if err := config.LoadDefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading default configuration: %s\n", err)
		os.Exit(1)
	}
	if err := config.ParseProgramArguments(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid program arguments: %s\n", err)
		os.Exit(1)
	}
	if err := config.SanityCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "pre-flight check failed: %s\n", err)
		os.Exit(1)
	}

	// Run discovery service
	if config.Discovery.DiscoveryAddress != "" {
		if err := RunDiscoveryService(config.Discovery.DiscoveryAddress, config.Discovery.DiscoveryToken); err != nil {
			fmt.Fprintf(os.Stderr, "discovery service error: %s\n", err)
			os.Exit(1)
		}
		log.Printf("openqa-agent discovery running: %s", config.Discovery.DiscoveryAddress)
	}

	// Run agent serial terminal
	if config.Serial.SerialPort != "" {
		if err := RunSerialTerminalAgent(config.Serial.SerialPort, config); err != nil {
			fmt.Fprintf(os.Stderr, "serial port error: %s\n", err)
			os.Exit(1)
		}
		log.Printf("openqa-agent running on serial port %s", config.Serial.SerialPort)
	}

	// Run agent webserver
	if config.Webserver.BindAddress != "" {
		http.Handle("GET /health", healthHandler())
		http.Handle("GET /status", healthHandler())
		http.Handle("GET /health.json", healthHandler())
		http.Handle("GET /status.json", healthHandler())
		http.Handle("POST /exec", checkTokenHandler(execHandler(config), config))
		http.Handle("GET /file", checkTokenHandler(getFileHandler(), config))
		http.Handle("POST /file", checkTokenHandler(putFileHandler(), config))
		log.Printf("openqa-agent running: %s", config.Webserver.BindAddress)
		go func() {
			log.Fatal(http.ListenAndServe(config.Webserver.BindAddress, nil))
		}()
	}

	if err := RunService(); err != nil {
		log.Fatalf("error running as service: %e", err)
		os.Exit(1)
	}
	awaitTerminationSignal()
	os.Exit(1)
}

// awaits SIGINT or SIGTERM
func awaitTerminationSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	fmt.Println(sig)
}
