package log

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func nextLine(buf *bytes.Buffer) (string, error) {
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}

	i := strings.Index(line, "[")
	if i < 0 {
		return "", errors.New("failed to find start of log message")
	}

	return strings.TrimSuffix(line[i:], "\n"), nil
}

func TestFormatLogLevel(t *testing.T) {
	if level := formatLogLevel(Trace); level != "[TRACE]" {
		t.Errorf("invalid log level: %v", level)
	}
	if level := formatLogLevel(Debug); level != "[DEBUG]" {
		t.Errorf("invalid log level: %v", level)
	}
	if level := formatLogLevel(Info); level != "[INFO]" {
		t.Errorf("invalid log level: %v", level)
	}
	if level := formatLogLevel(Warn); level != "[WARN]" {
		t.Errorf("invalid log level: %v", level)
	}
	if level := formatLogLevel(Error); level != "[ERROR]" {
		t.Errorf("invalid log level: %v", level)
	}
	if level := formatLogLevel(Fatal); level != "[FATAL]" {
		t.Errorf("invalid log level: %v", level)
	}
}

func TestParseLogLevel(t *testing.T) {
	level, err := ParseLogLevel("trace")
	if err != nil {
		t.Error(err)
	}
	if level != Trace {
		t.Errorf("invalid log level: %v", level)
	}

	level, err = ParseLogLevel("DEBUG")
	if err != nil {
		t.Error(err)
	}
	if level != Debug {
		t.Errorf("invalid log level: %v", level)
	}

	level, err = ParseLogLevel("Info")
	if err != nil {
		t.Error(err)
	}
	if level != Info {
		t.Errorf("invalid log level: %v", level)
	}

	level, err = ParseLogLevel("wArN")
	if err != nil {
		t.Error(err)
	}
	if level != Warn {
		t.Errorf("invalid log level: %v", level)
	}

	level, err = ParseLogLevel("ErRoR")
	if err != nil {
		t.Error(err)
	}
	if level != Error {
		t.Errorf("invalid log level: %v", level)
	}

	level, err = ParseLogLevel("fatal")
	if err != nil {
		t.Error(err)
	}
	if level != Fatal {
		t.Errorf("invalid log level: %v", level)
	}

	_, err = ParseLogLevel("")
	if err == nil {
		t.Error("ParseLogLevel should fail with empty string")
	}
}

func TestStandardLogger(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test that the default level is set to Warn
	logger := NewStandardLogger()
	logger.SetOutput(buf)
	logger.Print(Info, "Print")
	logger.Print(Warn, "Print")
	line, err := nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[WARN] polaris/log.TestStandardLogger Print" {
		t.Fatalf("%q", line)
	}

	// Fatal cannot be tested due to them aborting execution.
	logger.SetLogLevel(Info)
	logger.Print(Trace, "Print")
	logger.Print(Debug, "Print")
	logger.Print(Info, "Print")
	logger.Print(Warn, "Print")
	logger.Print(Error, "Print")

	line, err = nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[INFO] polaris/log.TestStandardLogger Print" {
		t.Fatalf("%q", line)
	}

	line, err = nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[WARN] polaris/log.TestStandardLogger Print" {
		t.Fatalf("%q", line)
	}

	line, err = nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[ERROR] polaris/log.TestStandardLogger Print" {
		t.Fatalf("%q", line)
	}

	// Fatal cannot be tested due to them aborting execution.
	logger.SetLogLevel(Warn)
	logger.Printf(Trace, "Printf %q", "trace")
	logger.Printf(Debug, "Printf %q", "debug")
	logger.Printf(Info, "Printf %q", "info")
	logger.Printf(Warn, "Printf %q", "warn")
	logger.Printf(Error, "Printf %q", "error")

	line, err = nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[WARN] polaris/log.TestStandardLogger Printf \"warn\"" {
		t.Fatalf("%q", line)
	}
	line, err = nextLine(buf)
	if err != nil {
		t.Fatal(err)
	}
	if line != "[ERROR] polaris/log.TestStandardLogger Printf \"error\"" {
		t.Fatalf("%q", line)
	}
}

func TestPkgFuncName(t *testing.T) {
	if pfn := PkgFuncName(1); pfn != "polaris/log.TestPkgFuncName" {
		t.Fatalf("invalid PkgFuncName: %v", pfn)
	}
}
