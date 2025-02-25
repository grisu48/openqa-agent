package main

// Config hold the global program configuration
type Config struct {
	Token       string // Authentication token
	BindAddress string // Address the webserver binds to
}

// Singleton program configuration
var config Config

func (cf *Config) SetDefaults() {
	cf.Token = ""
	cf.BindAddress = ":8421"
}
