//go:build windows
// +build windows

package main

import "os/exec"

func (job *ExecJob) applySystemSettings(cmd *exec.Cmd) {
	// Doesn't support setting any other user yet
}
