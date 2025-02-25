package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

const MAX_BUFFER = 1024 * 1024 * 64 // Maximum buffer size for stdout and stderr

type ExecJob struct {
	Command string   `json:"cmd"`     // Command to be executed
	UID     int      `json:"uid"`     // User ID of the command to be executed
	GID     int      `json:"gid"`     // Group ID of the command to be executed
	Timeout int64    `json:"timeout"` // Timeout in seconds until the command is abandoned
	Env     []string `json:"env"`     // Environment variables

	ret     int    // Return code of the job
	runtime int64  // Runtime of the command in milliseconds
	stdout  []byte // Filled with the contents of stdout once executed
	stderr  []byte // Filled with the contents of stderr once executed
}

func (job *ExecJob) SetDefaults() {
	job.UID = 0
	job.GID = 0
	job.Timeout = 30
	job.Env = make([]string, 0)
	job.ret = 0
	job.stdout = nil
	job.stderr = nil
}

func (job *ExecJob) SanityCheck() error {
	if job.Command == "" {
		return fmt.Errorf("no command")
	}
	if job.UID < 0 {
		return fmt.Errorf("invalid uid")
	}
	if job.GID < 0 {
		return fmt.Errorf("invalid gid")
	}
	if job.Timeout <= 0 {
		return fmt.Errorf("invalid timeout")
	}
	return nil
}

// exec runs the given command and returns its exit status.
func (job *ExecJob) exec() error {
	// Split command into arguments as expected by exec.Command
	split := CommandSplit(job.Command)
	args := make([]string, 0)
	if len(split) > 1 {
		args = split[1:]
	}
	cmd := exec.Command(split[0], args...)
	// Apply command settings
	if job.UID > 0 || job.GID > 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(job.UID), Gid: uint32(job.GID)}
	}
	cmd.Env = job.Env

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Connect stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go ReadPipe(stdoutPipe, &stdout, MAX_BUFFER)
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go ReadPipe(stderrPipe, &stderr, MAX_BUFFER)

	// Run command
	job.runtime = time.Now().UnixMilli()
	if err := cmd.Start(); err != nil {
		return err
	}

	running := true
	// Wait for job completion
	completed := make(chan error, 1)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(job.Timeout) * time.Second)
		timeout <- true
	}()
	go func() {
		err := cmd.Wait()
		completed <- err
	}()
	var ret error
	for running {
		select {
		case <-completed:
			running = false
			ret = nil // Don't tread failed commands as program errors
		case <-timeout:
			cmd.Process.Kill()
			ret = fmt.Errorf("command timeout")
			running = false
		}
	}

	// Collect stats
	job.runtime = time.Now().UnixMilli() - job.runtime
	job.stdout = stdout.Bytes()
	job.stderr = stderr.Bytes()
	job.ret = cmd.ProcessState.ExitCode()
	return ret
}
