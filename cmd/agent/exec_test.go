package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	// Execute command, return object
	exec := func(command string) (ExecJob, error) {
		var job ExecJob
		job.SetDefaults()
		job.Command = command
		return job, job.exec()
	}
	execShell := func(command string, shell string) (ExecJob, error) {
		var job ExecJob
		job.SetDefaults()
		job.Command = command
		job.Shell = shell
		return job, job.exec()
	}
	execCwd := func(command string, cwd string) (ExecJob, error) {
		var job ExecJob
		job.SetDefaults()
		job.Command = command
		job.WorkDir = cwd
		return job, job.exec()
	}
	execTimeout := func(command string, timeout int64) (ExecJob, error) {
		var job ExecJob
		job.SetDefaults()
		job.Command = command
		job.Timeout = timeout
		return job, job.exec()
	}

	// Assert normal execution works
	job, err := exec("true")
	assert.NoError(t, err, "execution of true should succeed")
	assert.Equal(t, job.ret, 0, "true should terminate with ret = 0")
	// Assert execution of commands with arguments work works
	job, err = exec("bash -c true")
	assert.NoError(t, err, "execution of bash-true should succeed")
	assert.Equal(t, job.ret, 0, "bash-true should terminate with ret = 0")
	// Assert failing command execution works
	job, err = exec("false")
	assert.NoError(t, err, "execution of false should succeed")
	assert.NotEqual(t, job.ret, 0, "false should terminate with ret != 0")
	// Assert execution in shell works
	job, err = execShell("true", "bash")
	assert.NoError(t, err, "execution of true in bash should succeed")
	assert.Equal(t, job.ret, 0, "true in bash should terminate with ret = 0")
	// Assert execution of complex commands in shell works
	job, err = execShell("true && true", "bash")
	assert.NoError(t, err, "execution of true&&true in bash should succeed")
	assert.Equal(t, job.ret, 0, "true&&true in bash should terminate with ret = 0")
	job, err = execShell("true && false", "bash")
	assert.NoError(t, err, "execution of true&&false in bash should succeed")
	assert.NotEqual(t, job.ret, 0, "true&&false in bash should terminate with ret = 0")
	// Assert stdout works
	job, err = exec("echo 'hello world'")
	assert.NoError(t, err, "execution of echo should succeed")
	assert.Equal(t, job.ret, 0, "echo should terminate with ret = 0")
	assert.Contains(t, string(job.stdout), "hello world", "stdout capture should work")
	// Assert stderr works
	job, err = execShell("echo 'hello world' 1>&2", "bash")
	assert.NoError(t, err, "execution of echo-to-stderr should succeed")
	assert.Equal(t, job.ret, 0, "echo-to-stderr should terminate with ret = 0")
	assert.Contains(t, string(job.stderr), "hello world", "stderr capture should work")
	assert.NotContains(t, string(job.stdout), "hello world", "stdout capture should not contain the output")
	// Assert cwd works
	job, err = execCwd("pwd", "/tmp")
	assert.NoError(t, err, "execution of pwd should succeed")
	assert.Equal(t, job.ret, 0, "pwd should terminate with ret = 0")
	assert.Contains(t, string(job.stdout), "/tmp", "pwd should run in /tmp")
	// Assert timeout works
	job, err = execTimeout("sleep 2", 5)
	assert.NoError(t, err, "execution of sleep should succeed")
	assert.Equal(t, job.ret, 0, "sleep should terminate with ret = 0")
	assert.LessOrEqual(t, job.runtime, int64(4000), "timeout test should take less than 4 seconds (took %d)", job.runtime)
	assert.GreaterOrEqual(t, job.runtime, int64(1000), "sleep 2 should take more than 1 second (took %d)", job.runtime)
	job, err = execTimeout("sleep 5", 2)
	assert.Error(t, err, "execution of sleep should fail")
	assert.ErrorIs(t, err, TimeoutError, "timeout should run into a TimeoutError")
	assert.NotEqual(t, job.ret, 0, "sleep 5 should run into a timeout")
	assert.GreaterOrEqual(t, job.runtime, int64(1900), "timeout(sleep 5, 2) should take more than 1.9 seconds (took %d)", job.runtime)
	assert.LessOrEqual(t, job.runtime, int64(4000), "timeout(sleep 5, 2) should take less than 4 seconds (took %d)", job.runtime)

}
