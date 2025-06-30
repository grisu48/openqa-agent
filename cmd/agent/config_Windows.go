//go:build windows
// +build windows

package main

const DEFAULT_CONFIG_PATH = "C:\\Program Files\\openqa-agent.yaml"

// Apply system-specific default settings, if any
func (cf *Config) SetSystemDefaults() {
	cf.DefaultShell = "powershell"
	cf.DefaultWorkDir = "C:\\"
}
