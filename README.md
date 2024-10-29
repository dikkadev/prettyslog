# Prettyslog

`prettyslog` is a custom Go `slog` handler that formats log records according to your preferences. It provides a clean, colored text output with configurable options, ensuring your logs are both informative and easy to read.

## Features

- **Customizable Log Format**: Define how your logs look, including timestamp formats and attribute styling.
- **Color Support**: Enhance readability with color-coded log levels and keys.
- **Group Handling**: Required group names at the start of each log line, with support for nested groups.
- **Functional Options**: Easily configure the handler using functional options.
- **Thread-Safe**: Safe for concurrent use by multiple goroutines.

## Installation
```bash
go get github.com/dikkadev/prettyslog
```

## Usage

```go
package main

import (
    "log/slog"
    "github.com/dikkadev/prettyslog"
)

func main() {
    // Create a new prettyslog handler with the group "MyApp" and default options
    handler := prettyslog.NewPrettyslogHandler("MyApp")

    // Set the default slog logger to use our custom handler
    slog.SetDefault(slog.New(handler))

    // Now, use slog as usual
    slog.Info("Application started", "version", "1.0.0")

    // Using a logger with additional attributes
    logger := slog.With("request_id", "12345")
    logger.Info("User logged in", "username", "johndoe")

    // Using a different group
    dbLogger := logger.WithGroup("db")
    dbLogger.Info("Query executed", "query", "SELECT * FROM users")

    // Including source information
    handlerWithSource := prettyslog.NewPrettyslogHandler("MyApp",
        prettyslog.WithSource(true),
    )
    slog.SetDefault(slog.New(handlerWithSource))
    slog.Info("This log includes source information")
}
```

## Configuration Options

`prettyslog` uses functional options for configuration:

- `WithTimestamp(enabled bool)`: Enable or disable timestamps. Default is `true`.
- `WithTimestampFormat(format string)`: Set a custom timestamp format. Default is `"[15:04:05.000]"`.
- `WithSource(enabled bool)`: Include source file and line number. Default is `false`.
- `WithColors(enabled bool)`: Enable or disable colors. Default is `true`.
- `WithLevel(level slog.Leveler)`: Set the minimum log level. Default is `slog.LevelInfo`.
- `WithWriter(w io.Writer)`: Set a custom `io.Writer`. Default is `os.Stderr`.

### Example with Custom Options

```go
handler := prettyslog.NewPrettyslogHandler("MyApp",
    prettyslog.WithSource(true),
    prettyslog.WithLevel(slog.LevelDebug),
    prettyslog.WithColors(false),
    prettyslog.WithTimestampFormat("[2006-01-02 15:04:05]"),
)
slog.SetDefault(slog.New(handler))
```


