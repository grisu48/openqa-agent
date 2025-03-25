package main

import (
	"flag"
	"fmt"
)

// Config hold the global program configuration
type Config struct {
	Webserver      Webserver // Webserver configuration
	Discovery      Discovery // Discovery configuration
	Serial         Serial    // Serial port configuration
	DefaultShell   string    // Optional argument to run each command in this shell by default
	DefaultWorkDir string    // Default work dir for commands to be executed

}

type Webserver struct {
	Token       []Token // Accepted authentication token
	BindAddress string  // Address the webserver binds to
}

type Discovery struct {
	DiscoveryAddress string // Address where the discovery service runs
	DiscoveryToken   string // Unique token for the discovery service, if present
}

type Serial struct {
	SerialPort string // Serial Port where the agent should run on. Format: DEVICE[:BAUD]
}

// Authentication token object
type Token struct {
	Token string // Actual secret
}

// Singleton program configuration
var config Config

func (cf *Config) SetDefaults() {
	cf.Webserver.Token = make([]Token, 0)
	cf.Webserver.BindAddress = ""
	cf.DefaultShell = ""
	cf.DefaultWorkDir = ""
	cf.Discovery.DiscoveryAddress = ""
	cf.Discovery.DiscoveryToken = ""
	cf.Serial.SerialPort = ""
}

// Parse program arguments and apply settings to the config instance
func (cf *Config) ParseProgramArguments() error {
	var token = flag.String("t", "", "authentication token")
	var bind = flag.String("b", "", "webserver server address")
	var shell = flag.String("s", "", "default shell")
	var workDir = flag.String("c", "", "default work dir")
	var discovery = flag.String("d", "", "server discovery address")
	var discoveryToken = flag.String("i", "", "discovery token")
	var serialPort = flag.String("p", "", "serial port (PORT[:BAUD,PARITY,DATABITS,STOPBITS])")
	flag.Parse()
	if *token != "" {
		cf.Webserver.Token = append(cf.Webserver.Token, Token{Token: *token})
	}
	if *bind != "" {
		cf.Webserver.BindAddress = *bind
	}
	if *shell != "" {
		cf.DefaultShell = *shell
	}
	if *workDir != "" {
		cf.DefaultWorkDir = *workDir
	}
	if *discovery != "" {
		cf.Discovery.DiscoveryAddress = *discovery
	}
	if *discoveryToken != "" {
		cf.Discovery.DiscoveryToken = *discoveryToken
	}
	if *serialPort != "" {
		cf.Serial.SerialPort = *serialPort
	}
	return nil
}

// Perform sanity checks on the config and return errors find
func (cf *Config) SanityCheck() error {
	if cf.Webserver.BindAddress != "" && len(cf.Webserver.Token) == 0 {
		return fmt.Errorf("no access tokens for webserver")
	}
	if cf.Webserver.BindAddress == "" && cf.Serial.SerialPort == "" {
		return fmt.Errorf("neither serial nor webserver defined")
	}
	return nil
}

// CheckToken checks if the given token is allowed by the configuration
func (cf *Config) CheckToken(token string) bool {
	if token == "" {
		return false
	}
	for _, tok := range cf.Webserver.Token {
		// Additional check: Do not ever allow an empty token, even accidentally
		if tok.Token != "" && tok.Token == token {
			return true
		}
	}
	return false
}
