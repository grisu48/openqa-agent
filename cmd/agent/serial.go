package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	sr "go.bug.st/serial"
)

type Reply struct {
	Command    string `json:"cmd"`     // Command that was executed
	Shell      string `json:"shell"`   // Optional shell in which the command was executed
	Runtime    int64  `json:"runtime"` // Command runtime
	ReturnCode int    `json:"ret"`     // Return code
	StdOut     string `json:"stdout"`  // Standard output
	StdErr     string `json:"stderr"`  // Standard error
}

// Parse the given serial port argument into port and mode
// Acceptable input is e.g. 'COM1,9600,None,8,one' or '/dev/ttyS0,115200,0,8,1'
func parseSerialPort(port string) (string, *sr.Mode, error) {
	var err error

	// Default settings
	mode := sr.Mode{
		BaudRate: 115200,
		Parity:   sr.NoParity,
		DataBits: 8,
		StopBits: sr.OneStopBit,
	}

	// Parse possible options
	if i := strings.Index(port, ":"); i >= 0 {
		if i == 0 {
			return port, &mode, fmt.Errorf("missing serial port")
		}
		// Missing arguments. Ignore but still crop the separator
		if i >= len(port)-1 {
			port = port[:i]
		} else {
			// Split options and port
			options := strings.Split(port[i+1:], ",")
			port = port[:i]

			// PORT,[BAUD,PARITY,DATABITS,STOPBITS]

			// Parse arguments. All arguments are optional
			if options[0] != "" { // Baud
				mode.BaudRate, err = strconv.Atoi(options[0])
				if err != nil {
					return port, &mode, fmt.Errorf("invalid baud rate: %s", err)
				}
			}
			if len(options) > 1 && options[1] != "" { // Parity
				switch strings.ToLower(options[1]) {
				case "no", "none":
					mode.Parity = sr.NoParity
				case "even":
					mode.Parity = sr.EvenParity
				case "odd":
					mode.Parity = sr.OddParity
				case "mark":
					mode.Parity = sr.MarkParity
				case "space":
					mode.Parity = sr.SpaceParity
				default:
					return port, &mode, fmt.Errorf("invalid parity: %s", err)
				}
			}
			if len(options) > 2 && options[2] != "" { // Databits
				mode.DataBits, err = strconv.Atoi(options[2])
				if err != nil {
					return port, &mode, fmt.Errorf("invalid databits: %s", err)
				}
			}
			if len(options) > 3 && options[3] != "" { // Stop bits
				switch strings.ToLower(options[3]) {
				case "1", "one":
					mode.StopBits = sr.OneStopBit
				case "1.5", "onepointfive":
					mode.StopBits = sr.OnePointFiveStopBits
				case "2", "two":
					mode.StopBits = sr.TwoStopBits
				default:
					return port, &mode, fmt.Errorf("invalid stop bits: %s", err)
				}
			}

		}
	}
	return port, &mode, nil
}

// RunSerialTerminalAgent runs the agent against a given serial port. Returns and error if occuring while connecting to the port
func RunSerialTerminalAgent(dest string, conf Config) error {
	// Default settings
	port, mode, err := parseSerialPort(dest)
	if err != nil {
		return err
	}
	serial, err := sr.Open(port, mode)
	if err != nil {
		return err
	}
	go func() {
		defer serial.Close()
		if err := runSerialTerminalAgent(serial, conf); err != nil {
			log.Fatalf("serial port error: %s", err)
		}
	}()
	return nil
}

// Reads from the given stream and will execute all commands.
func runSerialTerminalAgent(stream io.ReadWriter, conf Config) error {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if command == "" || len(command) < 1 {
			continue
		}
		if command[0] == ':' {
			// Reserved for special commands, not used currently
			continue
		}

		// By design, each command will get it's own fresh struct. This is to avoid possible carry-over of some properties.
		var job ExecJob
		job.SetDefaults()
		job.Shell = conf.DefaultShell
		job.WorkDir = conf.DefaultWorkDir
		job.Command = command
		job.Timeout = 60

		err := job.exec()

		var reply Reply
		reply.Command = command
		reply.Shell = job.Shell
		reply.Runtime = job.runtime
		reply.ReturnCode = job.ret
		reply.StdOut = string(job.stdout)
		reply.StdErr = string(job.stderr)
		if err != nil {
			if errors.Is(err, TimeoutError) {
				reply.ReturnCode = 124
			} else {
				log.Fatalf("execution of '%s' failed: %s", command, err)
				reply.ReturnCode = -1
			}
		}

		log.Printf("serial command: '%s' -> %d", command, reply.ReturnCode)
		if buf, err := json.Marshal(reply); err != nil {
			return err
		} else {
			if _, err := stream.Write(buf); err != nil {
				return err
			}
			// Add termination character to mark the end of the json object
			if conf.Serial.Serialized {
				if _, err := stream.Write([]byte{0}); err != nil {
					return err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
