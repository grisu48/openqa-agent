package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config hold the global program configuration
type Config struct {
	Webserver      Webserver `yaml:"webserver"` // Webserver configuration
	Discovery      Discovery `yaml:"discovery"` // Discovery configuration
	Serial         Serial    `yaml:"serial"`    // Serial port configuration
	DefaultShell   string    `yaml:"shell"`     // Optional argument to run each command in this shell by default
	DefaultWorkDir string    `yaml:"workdir"`   // Default work dir for commands to be executed

}

type Webserver struct {
	Token       []Token `yaml:"token"` // Accepted authentication token
	BindAddress string  `yaml:"bind"`  // Address the webserver binds to
}

type Discovery struct {
	DiscoveryAddress string `yaml:"bind"`  // Address where the discovery service runs
	DiscoveryToken   string `yaml:"token"` // Unique token for the discovery service, if present
}

type Serial struct {
	SerialPort string `yaml:"port"` // Serial Port where the agent should run on. Format: DEVICE[:BAUD]
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
	var yamlFile = flag.String("f", "", "yaml configuration file")
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
	// Note: Load yaml file must happen at last
	if *yamlFile != "" {
		if err := cf.LoadYaml(*yamlFile); err != nil {
			return err
		}
	}
	return nil
}

// Load settings from a yaml file
func (cf *Config) LoadYaml(filename string) error {
	f, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(f, cf)
}

func (cf *Config) LoadDefaultConfig() error {
	if _, err := os.Stat(DEFAULT_CONFIG_PATH); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return cf.LoadYaml(DEFAULT_CONFIG_PATH)
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
