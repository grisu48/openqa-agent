//go:build linux
// +build linux

package main

const DEFAULT_CONFIG_PATH = "/etc/openqa/openqa-agent.yaml"

// Apply system-specific default settings, if any
func (cf *Config) SetSystemDefaults() {
}
