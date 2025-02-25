//go:build linux
// +build linux

package main

import (
	"os/exec"
	"syscall"
)

func (job *ExecJob) applySystemSettings(cmd *exec.Cmd) {
	if job.UID > 0 || job.GID > 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(job.UID), Gid: uint32(job.GID)}
	}
}
