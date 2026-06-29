// Package logger provides a minimal structured logger that writes JSON lines
// to standard output. The package intentionally has zero external dependencies
// to keep the microservice's supply chain tight.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Level enumerates the supported log severities.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger emits structured JSON log lines to an underlying writer.
type Logger struct {
	mu  sync.Mutex
	out *os.File
}

// New constructs a Logger writing to stdout.
func New() *Logger {
	return &Logger{out: os.Stdout}
}

func (l *Logger) write(level Level, msg string, fields map[string]any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	record := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": string(level),
		"msg":   msg,
	}
	for k, v := range fields {
		record[k] = v
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		_, _ = fmt.Fprintf(l.out, "{\"ts\":\"%s\",\"level\":\"error\",\"msg\":\"log_encode_failed\",\"err\":\"%s\"}\n",
			time.Now().UTC().Format(time.RFC3339Nano), err.Error())
		return
	}
	if _, err := l.out.Write(append(encoded, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "log write failed: %s\n", err.Error())
	}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields map[string]any) { l.write(LevelDebug, msg, fields) }

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields map[string]any) { l.write(LevelInfo, msg, fields) }

// Warn logs a warn-level message.
func (l *Logger) Warn(msg string, fields map[string]any) { l.write(LevelWarn, msg, fields) }

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields map[string]any) { l.write(LevelError, msg, fields) }
