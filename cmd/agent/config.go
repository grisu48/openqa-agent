package main

import (
	"flag"
	"fmt"
)

// Config hold the global program configuration
type Config struct {
	Token       []Token // Accepted authentication token
	BindAddress string  // Address the webserver binds to
}

type Token struct {
	Token string // Actual secret
}

// Singleton program configuration
var config Config

func (cf *Config) SetDefaults() {
	cf.Token = make([]Token, 0)
	cf.BindAddress = "127.0.0.1:8421"
}

func (cf *Config) ParseProgramArguments() error {
	var token = flag.String("t", "", "authentication token")
	var bind = flag.String("b", cf.BindAddress, "webserver bind ")
	flag.Parse()
	if *token != "" {
		cf.Token = append(cf.Token, Token{Token: *token})
	}
	if *bind != "" {
		cf.BindAddress = *bind
	}
	return nil
}

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
