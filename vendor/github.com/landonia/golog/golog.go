// Copyright 2016 Landon Wainwright. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package golog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Define the logging levels
const (

	// OFF logging level
	OFF int = 0

	// FATAL logging level
	FATAL int = 1

	// ERROR logging level
	ERROR int = 2

	// WARN logging level
	WARN int = 3

	// INFO logging level
	INFO int = 4

	// DEBUG logging level
	DEBUG int = 5

	// TRACE logging level
	TRACE int = 6
)

// Colour allows colours to be defined
type Colour int

// Define some colours for the output text
const (

	// RED color
	RED Colour = 31

	// GREEN color
	GREEN Colour = 32

	// YELLOW color
	YELLOW Colour = 33

	// BLUE color
	BLUE Colour = 34

	// MAG color
	MAG Colour = 35

	// CYAN color
	CYAN Colour = 36
)

var (

	// logLevel is the current global logging level
	logLevel = INFO

	// OutputLog is the base logger and can be overwritten on a package level if required
	OutputLog = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// file to write the log to
	file *os.File
)

// LogLevel wwill set the log level to the specified level
// if the log level is not recogised it will return a false and default to INFO
func LogLevel(ll string) bool {
	switch strings.ToUpper(ll) {
	case "FATAL":
		logLevel = FATAL
	case "ERROR":
		logLevel = ERROR
	case "WARN":
		logLevel = WARN
	case "INFO":
		logLevel = INFO
	case "DEBUG":
		logLevel = DEBUG
	case "TRACE":
		logLevel = TRACE
	default:
		logLevel = INFO
		return false
	}
	return true
}

// OutputToFile will override the log from printing to stdout and instead print to the specified file
// An error will be returned if the file could not be opened or created
func OutputToFile(filename string) error {

	// Attempt to open the file (create if it does not exist) and put the file in append mode
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {

		// The log could not be found or created
		return err
	}

	// Set the output to the file
	file = f
	OutputLog.SetOutput(f)
	return nil
}

// Close will close the underlying file
func Close() error {
	if file != nil {
		return file.Close()
	}
	return nil
}

// GoLog is a wrapper
type GoLog struct {
	ns string // The namespace for this log
}

// New will return a new log for the particular namespace
func New(ns string) *GoLog {
	return &GoLog{ns}
}

// print message to standard out prefixed with date and time
func print(level int, ns, s string) {
	if logLevel >= level {
		l := "TRACE"
		switch level {
		case FATAL:
			l = "FATAL"
		case ERROR:
			l = "ERROR"
		case WARN:
			l = "WARN"
		case INFO:
			l = "INFO"
		case DEBUG:
			l = "DEBUG"
		}
		OutputLog.Print(fmt.Sprintf("[%-5s] [%s] %s", l, ns, s))
	}
}

// formatString will format the string to the correct interface
func formatString(format string, params ...interface{}) string {
	return fmt.Sprintf(format, params...)
}

// PrintColour prints coloured message
func (gl *GoLog) PrintColour(level int, s string, colour Colour) {
	print(level, gl.ns, fmt.Sprintf("\x1b[%v;1m%v\x1b[0m", colour, s))
}

// Fatal prints a Fatal level message
func (gl *GoLog) Fatal(format string, params ...interface{}) {
	gl.PrintColour(FATAL, formatString(format, params...), RED)
	os.Exit(1)
}

// Error prints an Error level message
func (gl *GoLog) Error(format string, params ...interface{}) {
	gl.PrintColour(ERROR, formatString(format, params...), RED)
}

// Warn prints a Warn level message
func (gl *GoLog) Warn(format string, params ...interface{}) {
	gl.PrintColour(WARN, formatString(format, params...), YELLOW)
}

// Info prints an Info level message
func (gl *GoLog) Info(format string, params ...interface{}) {
	gl.PrintColour(INFO, formatString(format, params...), GREEN)
}

// Debug prints a Debug level message
func (gl *GoLog) Debug(format string, params ...interface{}) {
	print(DEBUG, gl.ns, formatString(format, params...))
}

// Trace prints a Trace level message
func (gl *GoLog) Trace(format string, params ...interface{}) {
	print(TRACE, gl.ns, formatString(format, params...))
}
