// Copyright 2021 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package log contains the Logger interface used by the Polaris SDK. The
// interface can be used to implement adapters for existing log frameworks.
package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel specifies the severity of a log entry. The SDK uses 6 different
// log levels.
type LogLevel int

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

// formatLogLevel formats the given LogLevel as a string. If level is an
// invalid log level FATAL is returned.
func formatLogLevel(level LogLevel) string {
	switch level {
	case Trace:
		return "[TRACE]"
	case Debug:
		return "[DEBUG]"
	case Info:
		return "[INFO]"
	case Warn:
		return "[WARN]"
	case Error:
		return "[ERROR]"
	default:
		return "[FATAL]"
	}
}

// ParseLogLevel parses the given string as a LogLevel.
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "trace":
		return Trace, nil
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn":
		return Warn, nil
	case "error":
		return Error, nil
	case "fatal":
		return Fatal, nil
	default:
		return Trace, fmt.Errorf("invalid log level: %q", level)
	}
}

// Logger used by the SDK. The SDK provides two implementations: DiscardLogger
// and StandardLogger.
type Logger interface {
	// SetLogLevel sets the log level to the specified level.
	SetLogLevel(level LogLevel)

	// Print writes to the implementing logger.
	Print(level LogLevel, args ...any)

	// Printf writes to the implementing logger.
	Printf(level LogLevel, format string, args ...any)
}

// DiscardLogger discards everything written. Note that this logger never
// panics.
type DiscardLogger struct{}

// SetLogLevel discards the log level.
func (l DiscardLogger) SetLogLevel(level LogLevel) {
}

// Print discards the given arguments.
func (l DiscardLogger) Print(level LogLevel, args ...any) {
}

// Printf discards the given arguments.
func (l DiscardLogger) Printf(level LogLevel, format string, args ...any) {
}

// StandardLogger uses the standard logger from Go's log package. The Fatal
// log level maps to log.Fatal and all other log levels map to log.Print.
type StandardLogger struct {
	level LogLevel
	*log.Logger
}

// NewStandardLogger returns a standard logger with level set to Warn.
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{level: Warn, Logger: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)}
}

// SetLogLevel sets the log level to the specified level.
func (l *StandardLogger) SetLogLevel(level LogLevel) {
	l.level = level
}

// Print writes to the standard logger. Arguments are handled in the manner of
// fmt.Print.
func (l *StandardLogger) Print(level LogLevel, args ...any) {
	if level < l.level {
		return
	}

	pkgFuncName := " "
	if name := PkgFuncName(2); name != "" {
		pkgFuncName = " " + name + " "
	}

	args = append([]any{formatLogLevel(level), pkgFuncName}, args...)
	switch level {
	case Fatal:
		l.Fatal(args...)
	default:
		l.Logger.Print(args...)
	}
}

// Printf writes to the standard logger. Arguments are handled in the manner of
// fmt.Print.
func (l *StandardLogger) Printf(level LogLevel, format string, args ...any) {
	if level < l.level {
		return
	}

	pkgFuncName := " "
	if name := PkgFuncName(2); name != "" {
		pkgFuncName = " " + name + " "
	}

	args = append([]any{formatLogLevel(level), pkgFuncName}, args...)
	switch level {
	case Fatal:
		l.Fatalf("%s%s"+format, args...)
	default:
		l.Logger.Printf("%s%s"+format, args...)
	}
}
