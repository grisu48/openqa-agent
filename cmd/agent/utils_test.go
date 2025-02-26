package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandSplit(t *testing.T) {
	assert.ElementsMatch(t, CommandSplit("abc123"), []string{"abc123"}, "split command of single word shoud return a single word")
	assert.ElementsMatch(t, CommandSplit("abs 123"), []string{"abs", "123"}, "split command of two word shoud return two word")
	assert.ElementsMatch(t, CommandSplit("abs \"123 456\""), []string{"abs", "123 456"}, "split command of escaped sequence must pass")
	assert.ElementsMatch(t, CommandSplit("abs '123 456'"), []string{"abs", "123 456"}, "split command of escaped sequence must pass")
	assert.ElementsMatch(t, CommandSplit("abs '123 \"456 789\"' abc"), []string{"abs", "123 \"456 789\"", "abc"}, "split command of escaped sequence must pass")
	assert.ElementsMatch(t, CommandSplit("abs \"123 '456 789'\" abc"), []string{"abs", "123 '456 789'", "abc"}, "split command of escaped sequence must pass")
}

func TestReadPipe(t *testing.T) {
	TEST_STRING := "abcdefg hijklmn opqrst uvw xyz 123 456 7890"
	reader := io.NopCloser(strings.NewReader(TEST_STRING))
	var buffer bytes.Buffer

	// Test ReadPipe
	err := ReadPipe(reader, &buffer, len(TEST_STRING))
	assert.ErrorIs(t, err, io.EOF, "ReadPipe should return EOF")
	assert.Equal(t, buffer.String(), TEST_STRING, "ReadPipe should read string")

	// Test limit function
	CROPPED := TEST_STRING[:10]
	reader = io.NopCloser(strings.NewReader(TEST_STRING)) // Require a fresh reader
	buffer.Reset()
	err = ReadPipe(reader, &buffer, len(CROPPED))
	assert.NoError(t, err, "ReadPipe should pass")
	assert.Equal(t, buffer.String(), CROPPED, "ReadPipe should read string")
}
