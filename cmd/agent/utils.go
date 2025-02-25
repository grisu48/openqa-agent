package main

import (
	"bytes"
	"io"
)

// CommandSplit splits a command into program arguments and obeys quotation marks
func CommandSplit(command string) []string {
	null := rune(0)
	esc := null // Escape character or \0 if not escaped currently
	ret := make([]string, 0)
	buf := "" // Current command

	for _, char := range command {
		if esc != null {
			if char == esc {
				esc = null
			} else {
				buf += string(char)
			}
		} else {
			// Check for quotation marks
			if char == '\'' || char == '"' {
				esc = char
			} else if char == ' ' {
				ret = append(ret, buf)
				buf = ""
			} else {
				buf += string(char)
			}
		}
	}
	// Remaining characters
	if buf != "" {
		ret = append(ret, buf)
		buf = ""
	}
	return ret
}

// ReadPipe reads from the given reader up until limit bytes. The maximum limit is not pre-allocated to allow a large maximum while not wasting memory unless necessary
// This routine is intended to use as reader from stdout and stderr
func ReadPipe(reader io.ReadCloser, writer *bytes.Buffer, limit int) error {
	buf := make([]byte, 1024)
	for {
		if n, err := reader.Read(buf); err != nil {
			reader.Close()
			return err
		} else if n > 0 {
			if writer.Len()+n > limit {
				remaining := limit - writer.Len()
				writer.Write(buf[:remaining])
				return reader.Close()
			}
			writer.Write(buf[:n])
		}
	}
}
