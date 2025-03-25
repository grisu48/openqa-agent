package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	assert.Empty(t, cf.Webserver.Token, "tokens must be empty")
	assert.Empty(t, cf.DefaultShell, "default shell must be empty")
	assert.Empty(t, cf.DefaultWorkDir, "default work dir must be empty")
	assert.Empty(t, cf.Webserver.BindAddress, "BindAddress should be empty by default")
	assert.Empty(t, cf.Serial.SerialPort, "SerialPort should be empty by default")
}

func TestTokens(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	cf.Webserver.Token = append(cf.Webserver.Token, Token{Token: "secret"})
	cf.Webserver.Token = append(cf.Webserver.Token, Token{Token: "secret2"})
	assert.False(t, cf.CheckToken(""), "empty token must be rejected")
	assert.False(t, cf.CheckToken("wrong"), "empty token must be rejected")
	assert.True(t, cf.CheckToken("secret"), "valid token must be accepted")
	assert.True(t, cf.CheckToken("secret2"), "valid token must be accepted")
}

func TestSanityCheck(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	cf.Webserver.BindAddress = ""
	cf.Serial.SerialPort = ""
	assert.Error(t, cf.SanityCheck(), "sanity check must fail with no bind address and no serial port")
	cf.Serial.SerialPort = "/dev/ttyS0:115200"
	assert.NoError(t, cf.SanityCheck(), "sanity checks should pass with serial port on")
	cf.Serial.SerialPort = ""
	cf.Webserver.BindAddress = "127.0.0.1:8421"
	assert.Error(t, cf.SanityCheck(), "sanity check must fail with bind address and no tokens")
	cf.Webserver.Token = []Token{Token{Token: "nots3cr3t"}}
	assert.NoError(t, cf.SanityCheck(), "sanity checks should pass with webserver on")
	cf.Serial.SerialPort = "/dev/ttyS0:115200"
	assert.NoError(t, cf.SanityCheck(), "sanity checks should pass with webserver and serial port on")
}

func TestYaml(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	assert.NoError(t, cf.LoadYaml("../../config.yaml"), "loading example yaml should succeed")
	// Check applied settings
	assert.Equal(t, "bash", cf.DefaultShell)
	assert.Equal(t, "/home/geekotest", cf.DefaultWorkDir)
	// Webserver
	assert.Equal(t, ":8421", cf.Webserver.BindAddress)
	// Webserver tokens
	assert.NotEmpty(t, cf.Webserver.Token)
	assert.True(t, cf.CheckToken("nots3cr3t"), "first token should work")
	assert.True(t, cf.CheckToken("passw0rd"), "second token should work")
	assert.False(t, cf.CheckToken("password"), "wrong token should not work")
	// Discovery
	assert.Equal(t, ":8421", cf.Discovery.DiscoveryAddress)
	assert.Equal(t, "openqa-agent-1", cf.Discovery.DiscoveryToken)
	// Serial port
	assert.Equal(t, "/dev/ttyS0,115200", cf.Serial.SerialPort)
}
