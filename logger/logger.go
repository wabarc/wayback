// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package logger // import "github.com/wabarc/wayback/logger"

import (
	"fmt"
	"os"
	"time"
)

var logLevel = LevelInfo
var showTime = true

// LogLevel type.
type LogLevel uint32

const (
	// LevelFatal should be used in fatal situations, the app will exit.
	LevelFatal LogLevel = iota

	// LevelError should be used when someone should really look at the error.
	LevelError

	// LevelInfo should be used during normal operations.
	LevelInfo

	// LevelDebug should be used only during development.
	LevelDebug
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// DisableTime hides time in log messages.
func DisableTime() {
	showTime = false
}

// EnableDebug increases logging, more verbose (debug)
func EnableDebug() {
	logLevel = LevelDebug
	logging(LevelInfo, "Debug mode enabled")
}

// Debug sends a debug log message.
func Debug(format string, v ...interface{}) {
	if logLevel >= LevelDebug {
		logging(LevelDebug, format, v...)
	}
}

// Info sends an info log message.
func Info(format string, v ...interface{}) {
	if logLevel >= LevelInfo {
		logging(LevelInfo, format, v...)
	}
}

// Error sends an error log message.
func Error(format string, v ...interface{}) {
	if logLevel >= LevelError {
		logging(LevelError, format, v...)
	}
}

// Fatal sends a fatal log message and stop the execution of the program.
func Fatal(format string, v ...interface{}) {
	if logLevel >= LevelFatal {
		logging(LevelFatal, format, v...)
		os.Exit(1)
	}
}

func logging(l LogLevel, format string, v ...interface{}) {
	var prefix string

	if showTime {
		prefix = fmt.Sprintf("[%s] [%s] ", time.Now().Format("2006-01-02T15:04:05"), l)
	} else {
		prefix = fmt.Sprintf("[%s] ", l)
	}

	fmt.Fprintf(os.Stderr, prefix+format+"\n", v...)
}
