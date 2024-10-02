package prettyslog_test

import (
	"bytes"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/sett17/prettyslog"
)

func TestPrettyslogHandlerVisually(t *testing.T) {
	var buf bytes.Buffer

	// Create a new handler with default options
	handler := prettyslog.NewPrettyslogHandler("TestApp",
		prettyslog.WithWriter(&buf),
		prettyslog.WithLevel(slog.LevelDebug),
	)

	// Set the default logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log messages with various levels and attributes
	slog.Info("Application started", "version", "1.0.0")

	logger.Debug("Debug message", "debug_key", "debug_value")
	logger.Info("Info message", "info_key", "info_value")
	logger.Warn("Warning message", "warn_key", "warn_value")
	logger.Error("Error message", "error_key", "error_value")

	// Log with additional attributes
	logger = logger.With("session_id", "abc123")
	logger.Info("User logged in", "username", "johndoe")

	// Log with group
	logger.WithGroup("db").Info("Query executed", "query", "SELECT * FROM users")

	// Log with source information
	handlerWithSource := prettyslog.NewPrettyslogHandler("TestApp",
		prettyslog.WithWriter(&buf),
		prettyslog.WithSource(true),
	)
	loggerWithSource := slog.New(handlerWithSource)
	loggerWithSource.Info("This log includes source information")

	// Log without timestamp
	handlerWithoutTimestamp := prettyslog.NewPrettyslogHandler("TestApp",
		prettyslog.WithWriter(&buf),
		prettyslog.WithTimestamp(false),
	)
	loggerWithoutTimestamp := slog.New(handlerWithoutTimestamp)

	loggerWithoutTimestamp.Info("This log does not include a timestamp")

	// Log without colors
	handlerWithoutColors := prettyslog.NewPrettyslogHandler("TestApp",
		prettyslog.WithWriter(&buf),
		prettyslog.WithColors(false),
	)
	loggerWithoutColors := slog.New(handlerWithoutColors)

	loggerWithoutColors.Info("This log does not include colors")

	// Log with custom timestamp format with date
	handlerWithCustomTimestamp := prettyslog.NewPrettyslogHandler("TestApp",
		prettyslog.WithWriter(&buf),
		prettyslog.WithTimestampFormat("<2006-01-02 15:04:05.000>"),
	)
	loggerWithCustomTimestamp := slog.New(handlerWithCustomTimestamp)

	loggerWithCustomTimestamp.Info("This log includes a custom timestamp format")

	// Output the buffer to standard output for demonstration purposes
	t.Log("\n" + buf.String())
}
func TestPrettyslogHandler(t *testing.T) {
	tests := []struct {
		name           string
		handlerOptions []prettyslog.Option
		logFunc        func(logger *slog.Logger)
		expected       string
	}{
		{
			name:           "DefaultOptions",
			handlerOptions: []prettyslog.Option{},
			logFunc: func(logger *slog.Logger) {
				logger.Info("Test message", "key", "value")
			},
			expected: `TestApp INF \[\d{2}:\d{2}:\d{2}\.\d{3}\] Test message; key=value`,
		},
		{
			name:           "WithGroup",
			handlerOptions: []prettyslog.Option{},
			logFunc: func(logger *slog.Logger) {
				logger = logger.WithGroup("db")
				logger.Info("DB message", "query", "SELECT * FROM users")
			},
			expected: `db INF \[\d{2}:\d{2}:\d{2}\.\d{3}\] DB message; query=SELECT \* FROM users`,
		},
		{
			name: "WithSource",
			handlerOptions: []prettyslog.Option{
				prettyslog.WithSource(true),
			},
			logFunc: func(logger *slog.Logger) {
				logger.Info("Source message")
			},
			expected: `TestApp INF \[\d{2}:\d{2}:\d{2}\.\d{3}\] .+\.go:\d+ Source message`,
		},
		{
			name:           "WithAttrs",
			handlerOptions: []prettyslog.Option{},
			logFunc: func(logger *slog.Logger) {
				logger = logger.With("session_id", "abc123")
				logger.Info("Session message", "user", "johndoe")
			},
			expected: `TestApp INF \[\d{2}:\d{2}:\d{2}\.\d{3}\] Session message; session_id=abc123; user=johndoe`,
		},
		{
			name: "WithoutTimestamp",
			handlerOptions: []prettyslog.Option{
				prettyslog.WithTimestamp(false),
			},
			logFunc: func(logger *slog.Logger) {
				logger.Info("No timestamp message")
			},
			expected: `TestApp INF No timestamp message`,
		},
		{
			name: "WithoutColors",
			handlerOptions: []prettyslog.Option{
				prettyslog.WithColors(false),
			},
			logFunc: func(logger *slog.Logger) {
				logger.Info("No colors message")
			},
			expected: `TestApp INF \[\d{2}:\d{2}:\d{2}\.\d{3}\] No colors message`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			// Append buffer to handler options
			options := append(tt.handlerOptions, prettyslog.WithWriter(&buf))
			handler := prettyslog.NewPrettyslogHandler("TestApp", options...)

			logger := slog.New(handler)
			tt.logFunc(logger)

			output := buf.String()

			// Remove ANSI escape codes for testing
			output = stripAnsi(output)
			output = strings.TrimSpace(output)

			// Use regex to match the expected pattern
			matched, err := regexp.MatchString("^"+tt.expected+"$", output)
			if err != nil {
				t.Fatalf("Error compiling regex: %v", err)
			}
			if !matched {
				t.Errorf("Expected output to match %q, got %q", tt.expected, output)
			}
		})
	}
}

func stripAnsi(str string) string {
	// Remove ANSI escape codes
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(str, "")
}
