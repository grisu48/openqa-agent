package main

import (
	"encoding/json"
	"log"
	"net"
)

// Runs the discovery service on the given address.
func RunDiscoveryService(address string, token string) error {
	server, err := net.ListenPacket("udp", address)
	if err != nil {
		return err
	}

	// Discovery response
	type Discovery struct {
		Agent  string `json:"agent"`
		Status string `json:"status"`
		Token  string `json:"token"`
	}
	var discover Discovery
	discover.Agent = "openqa-agent"
	discover.Status = "ok"
	discover.Token = token
	response, err := json.Marshal(discover)
	if err != nil {
		return err
	}
	go func() {
		defer server.Close()

		for {
			buf := make([]byte, 1500)
			_, addr, err := server.ReadFrom(buf)
			if err != nil {
				log.Fatalf("error receiving discovery packet: %s", err)
				return
			}
			if _, err := server.WriteTo(response, addr); err != nil {
				log.Fatalf("error sending discovery packet: %s", err)
			}
		}
	}()
	return nil
}
