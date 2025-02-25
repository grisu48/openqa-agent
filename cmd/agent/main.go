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
	if err := config.ParseProgramArguments(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid program arguments: %s\n", err)
		os.Exit(1)
	}
	if err := config.SanityCheck(); err != nil {
		fmt.Fprintf(os.Stderr, "pre-flight check failed: %s\n", err)
		os.Exit(1)
	}

	// Run agent webserver
	awaitTerminationSignal()
	http.Handle("GET /health", healthHandler())
	http.Handle("GET /status", healthHandler())
	http.Handle("GET /health.json", healthHandler())
	http.Handle("GET /status.json", healthHandler())
	http.Handle("POST /exec", checkTokenHandler(execHandler()))
	http.Handle("GET /file", checkTokenHandler(getFileHandler()))
	http.Handle("POST /file", checkTokenHandler(putFileHandler()))
	log.Printf("Listening on %s", config.BindAddress)
	log.Fatal(http.ListenAndServe(config.BindAddress, nil))
}

// awaits SIGINT or SIGTERM
func awaitTerminationSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		os.Exit(1)
	}()
}
