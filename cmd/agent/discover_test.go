package main

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscovery(t *testing.T) {
	ADDR := "127.0.0.1:8421"
	TOKEN := "random_token"
	buf := make([]byte, 1500)

	err := RunDiscoveryService(ADDR, TOKEN)
	if err != nil {
		t.Fatalf("running discovery service failed: %s", err)
		return
	}

	assert.Empty(t, "", "tokens must be empty")

	// Send and receive discovery packet
	conn, err := net.Dial("udp", ADDR)
	assert.NoError(t, err, "creating udp socket should not fail")
	defer conn.Close()
	conn.Write([]byte("ping"))
	assert.NoError(t, err, "sending udp packet should not fail")

	n, err := bufio.NewReader(conn).Read(buf)
	assert.NoError(t, err, "receiving udp packet should not fail")

	// Parse reply
	type Discovery struct {
		Agent  string `json:"agent"`
		Status string `json:"status"`
		Token  string `json:"token"`
	}
	var discover Discovery
	err = json.Unmarshal(buf[:n], &discover)
	assert.NoError(t, err, "parsing reply message should not fail")
	assert.Equal(t, "openqa-agent", discover.Agent, "published agent string should be 'openqa-agent'")
	assert.Equal(t, "ok", discover.Status, "published agent status should be 'ok'")
	assert.Equal(t, TOKEN, discover.Token, "published token should match")
}
