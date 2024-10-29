package setdefault

import (
	"log/slog"

	"github.com/dikkadev/prettyslog"
)

func init() {
	handler := prettyslog.NewPrettyslogHandler("app",
		prettyslog.WithSource(false),
		prettyslog.WithLevel(slog.LevelDebug),
		prettyslog.WithColors(true),
	)
	slog.SetDefault(slog.New(handler))
}
