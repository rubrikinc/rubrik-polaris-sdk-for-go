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

package polaris

import (
	"errors"
	"log"
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
		return "TRACE"
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "FATAL"
	}
}

// parseLogLevel parses the given string as a LogLevel.
func parseLogLevel(level string) (LogLevel, error) {
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
		return Trace, errors.New("polaris: invalid log level")
	}
}

// Logger used by the SDK. Specify the implementation to use when creating the
// Client with NewClient. The SDK provides two implementations: DiscardLogger
// and StandardLogger.
type Logger interface {
	// SetLogLevel sets the log level to the specified level.
	SetLogLevel(level LogLevel)

	// Print writes to the implementing logger.
	Print(level LogLevel, args ...interface{})

	// Printf writes to the implementing logger.
	Printf(level LogLevel, format string, args ...interface{})
}

// DiscardLogger discards everything written. Note that this logger never
// panics. When Config.Log is set to false the logger given to NewClient is
// silently replaced by a DiscardLogger.
type DiscardLogger struct{}

// SetLogLevel discards the log level.
func (l DiscardLogger) SetLogLevel(level LogLevel) {
}

// Print discards the specified arguments.
func (l DiscardLogger) Print(level LogLevel, args ...interface{}) {
}

// Printf discards the specified arguments.
func (l DiscardLogger) Printf(level LogLevel, format string, args ...interface{}) {
}

// StandardLogger uses the standard logger from Golang's log package. The Fatal
// log level maps to log.Fatal, the Error log level maps to log.Panic and all
// other log levels map to log.Print.
type StandardLogger struct {
	level LogLevel
}

// SetLogLevel sets the log level to the specified level.
func (l StandardLogger) SetLogLevel(level LogLevel) {
	l.level = level
}

// Print writes to the standard logger. Arguments are handled in the manner of
// fmt.Print.
func (l StandardLogger) Print(level LogLevel, args ...interface{}) {
	if level < l.level {
		return
	}

	args = append([]interface{}{formatLogLevel(level), " - "}, args...)
	switch level {
	case Fatal:
		log.Fatal(args...)

	case Error:
		log.Panic(args...)

	default:
		log.Print(args...)
	}
}

// Print writes to the standard logger. Arguments are handled in the manner of
// fmt.Print.
func (l StandardLogger) Printf(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	format = "%s - " + format
	args = append([]interface{}{formatLogLevel(level)}, args...)
	switch level {
	case Fatal:
		log.Fatalf(format, args...)

	case Error:
		log.Panicf(format, args...)

	default:
		log.Printf(format, args...)
	}
}
