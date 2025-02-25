package main

import (
	"flag"
	"fmt"
)

const DEFAULT_BIND_ADDRESS = "127.0.0.1:8421"

// Config hold the global program configuration
type Config struct {
	Token          []Token // Accepted authentication token
	BindAddress    string  // Address the webserver binds to
	DefaultShell   string  // Optional argument to run each command in this shell by default
	DefaultWorkDir string  // Default work dir for commands to be executed
}

// Authentication token object
type Token struct {
	Token string // Actual secret
}

// Singleton program configuration
var config Config

func (cf *Config) SetDefaults() {
	cf.Token = make([]Token, 0)
	cf.BindAddress = DEFAULT_BIND_ADDRESS
	cf.DefaultShell = ""
	cf.DefaultWorkDir = ""
}

// Parse program arguments and apply settings to the config instance
func (cf *Config) ParseProgramArguments() error {
	var token = flag.String("t", "", "authentication token")
	var bind = flag.String("b", DEFAULT_BIND_ADDRESS, "webserver bind ")
	var shell = flag.String("s", "", "default shell")
	var workDir = flag.String("c", "", "default work dir")
	flag.Parse()
	if *token != "" {
		cf.Token = append(cf.Token, Token{Token: *token})
	}
	if *bind != "" {
		cf.BindAddress = *bind
	}
	if *shell != "" {
		cf.DefaultShell = *shell
	}
	if *workDir != "" {
		cf.DefaultWorkDir = *workDir
	}
	return nil
}

// Perform sanity checks on the config and return errors find
func (cf *Config) SanityCheck() error {
	if len(cf.Token) == 0 {
		return fmt.Errorf("no access tokens")
	}
	return nil
}

// CheckToken checks if the given token is allowed by the configuration
func (cf *Config) CheckToken(token string) bool {
	if token == "" {
		return false
	}
	for _, tok := range cf.Token {
		// Additional check: Do not ever allow an empty token, even accidentally
		if tok.Token != "" && tok.Token == token {
			return true
		}
	}
	return false
}
