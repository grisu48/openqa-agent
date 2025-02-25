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
	config.SetDefaults()

	awaitTerminationSignal()
	http.HandleFunc("/health", createHealthHandler())
	http.HandleFunc("/health.json", createHealthHandler())
	http.HandleFunc("/exec", createExecHandler())
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
