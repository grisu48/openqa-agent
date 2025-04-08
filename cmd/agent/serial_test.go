package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	sr "go.bug.st/serial"
)

type TerminalEmulator struct {
	in  *bytes.Buffer // Input buffer
	out *bytes.Buffer // Output buffer
}

func (term *TerminalEmulator) Read(buf []byte) (int, error) {
	return term.in.Read(buf)
}

func (term *TerminalEmulator) Write(buf []byte) (int, error) {
	return term.out.Write(buf)
}

func (term *TerminalEmulator) Clear() {
	term.in.Reset()
	term.out.Reset()
}

func NewTerminalEmulator() TerminalEmulator {
	var term TerminalEmulator
	term.in = bytes.NewBuffer(nil)
	term.out = bytes.NewBuffer(nil)
	return term
}

func TestSerialTerminal(t *testing.T) {
	var conf Config
	var reply Reply
	terminal := NewTerminalEmulator()
	decoder := json.NewDecoder(terminal.out)

	conf.DefaultShell = "bash"
	conf.DefaultWorkDir = ""
	conf.Serial.Serialized = false

	// Run single command
	terminal.in.Write([]byte("true\n"))
	runSerialTerminalAgent(&terminal, conf)
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for `true` should be 0")

	terminal.Clear()
	decoder = json.NewDecoder(terminal.out)

	// Run multiple commands
	terminal.in.Write([]byte("! false\n"))
	terminal.in.Write([]byte("echo 1\n"))
	terminal.in.Write([]byte("sleep 1\n"))
	terminal.in.Write([]byte("false\n"))
	runSerialTerminalAgent(&terminal, conf)
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for `!false` should be 0")
	assert.Equal(t, "! false", reply.Command, "command should be `!false`")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for `echo 1` should be 0")
	assert.Equal(t, "echo 1", reply.Command, "command should be `echo 1`")
	assert.Equal(t, "1\n", reply.StdOut, "command stdout should be `1`")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for `sleep 1` should be 0")
	assert.Equal(t, "sleep 1", reply.Command, "command should be `sleep 1`")
	assert.GreaterOrEqual(t, reply.Runtime, int64(999), "runtime for `sleep 1` should be >= 1 second")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.NotEqual(t, 0, reply.ReturnCode, "return code for `false` should not be 0")
	assert.Equal(t, "false", reply.Command, "command should be `false`")

	terminal.Clear()
	decoder = json.NewDecoder(terminal.out)

	// Check if json command parsing works
	terminal.in.Write([]byte("{\"cmd\":\"true\"}\n"))
	terminal.in.Write([]byte("{\"cmd\":\"true\",\"shell\":\"bash\"}\n"))
	terminal.in.Write([]byte("{\"cmd\":\"sleep 1\"}\n"))
	terminal.in.Write([]byte("{\"cmd\":\"\"}\n"))
	runSerialTerminalAgent(&terminal, conf)
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for json-encoded `true` should be 0")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for json-encoded `true` should be 0")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.Equal(t, 0, reply.ReturnCode, "return code for `sleep 1` should be 0")
	assert.Equal(t, "sleep 1", reply.Command, "command should be `sleep 1`")
	assert.NoError(t, decoder.Decode(&reply), "reply parsing should succeed")
	assert.NotEqual(t, 0, reply.ReturnCode, "return code for (empty command) should not be 0")

}

func TestSerialTerminalParsing(t *testing.T) {
	assertMode := func(mode *sr.Mode, ref *sr.Mode) {
		assert.Equal(t, ref.BaudRate, mode.BaudRate, "baud rates should match")
		assert.Equal(t, ref.Parity, mode.Parity, "parity should match")
		assert.Equal(t, ref.DataBits, mode.DataBits, "data bits should match")
		assert.Equal(t, ref.StopBits, mode.StopBits, "stop bits should match")
	}

	// Expected mode. Start with default values
	defaults := sr.Mode{
		BaudRate: 115200,
		Parity:   sr.NoParity,
		DataBits: 8,
		StopBits: sr.OneStopBit,
	}
	expected := defaults

	// Check default handling for Windows and Linux ports
	port, mode, err := parseSerialPort("COM1")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("/dev/ttyS0")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "/dev/ttyS0", "/dev/ttyS0 should be parsed as /dev/ttyS0")
	assertMode(mode, &expected)

	// Check program argument handling. All arguments are optional
	port, mode, err = parseSerialPort("COM1:9600")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected.BaudRate = 9600
	assertMode(mode, &expected)

	// Test if all arguments are parsed
	port, mode, err = parseSerialPort("/dev/ttyS0:57600,even,7,2")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "/dev/ttyS0", "/dev/ttyS0 should be parsed as /dev/ttyS0")
	expected.BaudRate = 57600
	expected.Parity = sr.EvenParity
	expected.DataBits = 7
	expected.StopBits = sr.TwoStopBits
	assertMode(mode, &expected)

	// Test Windows-like arguments
	port, mode, err = parseSerialPort("COM1:9600,None,8,one")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected.BaudRate = 9600
	expected.Parity = 0
	expected.DataBits = 8
	expected.StopBits = sr.OneStopBit
	assertMode(mode, &expected)

	// Test partial argument parsing
	port, mode, err = parseSerialPort("COM1:,odd")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.Parity = sr.OddParity
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("COM1:9600,,,")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.BaudRate = 9600
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("COM1:9600,,,two")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.BaudRate = 9600
	expected.StopBits = 2
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("COM1:,,7")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.DataBits = 7
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("COM1:,,,2")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.StopBits = sr.TwoStopBits
	assertMode(mode, &expected)
	port, mode, err = parseSerialPort("COM1:,,,1.5")
	assert.NoError(t, err, "parseSerialPort should succeed")
	assert.Equal(t, port, "COM1", "COM1 should be parsed as COM1")
	expected = defaults
	expected.StopBits = sr.OnePointFiveStopBits
	assertMode(mode, &expected)
}

func TestSerialization(t *testing.T) {
	var conf Config
	var reply Reply
	terminal := NewTerminalEmulator()

	conf.DefaultShell = "bash"
	conf.DefaultWorkDir = ""
	conf.Serial.Serialized = true

	// Run command 10 times to ensure the serialization can properly distinguish between runs
	for range 10 {
		terminal.in.Write([]byte("true\000"))
		runSerialTerminalAgent(&terminal, conf)
		buf, err := terminal.out.ReadBytes(0)
		assert.NoError(t, err, "reading until termination symbol should succeed")
		assert.Greater(t, len(buf), 1, "reply buffer should be larger than 1 character")
		buf = buf[:len(buf)-1] // Remove termination character
		assert.NoError(t, json.Unmarshal(buf, &reply))
		assert.Equal(t, 0, reply.ReturnCode, "echo command should succeed")
	}
}
