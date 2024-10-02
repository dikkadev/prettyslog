package prettyslog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"log/slog"
)

const (
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"
	FgGray    = "\033[90m"

	ColorReset = "\033[0m"
)

type PrettyslogHandler struct {
	mu     *sync.Mutex
	writer io.Writer
	groups []string
	opts   options
	attrs  []slog.Attr
}

type options struct {
	includeTimestamp bool
	includeSource    bool
	colors           bool
	levelColors      map[slog.Level]string
	timestampFormat  string
	level            slog.Leveler
	writer           io.Writer
}

type Option func(*options)

func defaultOptions() options {
	return options{
		includeTimestamp: true,
		includeSource:    false,
		colors:           true,
		timestampFormat:  "[15:04:05.000]",
		levelColors: map[slog.Level]string{
			slog.LevelDebug: FgMagenta,
			slog.LevelInfo:  FgBlue,
			slog.LevelWarn:  FgYellow,
			slog.LevelError: FgRed,
		},
		level:  slog.LevelInfo,
		writer: os.Stderr,
	}
}

func WithTimestamp(enabled bool) Option {
	return func(o *options) {
		o.includeTimestamp = enabled
	}
}

func WithSource(enabled bool) Option {
	return func(o *options) {
		o.includeSource = enabled
	}
}

func WithColors(enabled bool) Option {
	return func(o *options) {
		o.colors = enabled
	}
}

func WithLevel(level slog.Leveler) Option {
	return func(o *options) {
		o.level = level
	}
}

func WithTimestampFormat(format string) Option {
	return func(o *options) {
		o.timestampFormat = format
	}
}

func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

func NewPrettyslogHandler(group string, opts ...Option) *PrettyslogHandler {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	handler := &PrettyslogHandler{
		mu:     &sync.Mutex{},
		writer: options.writer,
		groups: []string{group},
		opts:   options,
	}

	return handler
}

func (h *PrettyslogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.level.Level()
}

func (h *PrettyslogHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(nil, r.Level) {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var b strings.Builder

	groupName := h.groups[len(h.groups)-1]
	b.WriteString(h.colorize(groupName, h.opts.colors, ""))
	b.WriteString(" ")

	levelStr := h.formatLevel(r.Level)
	b.WriteString(levelStr)
	b.WriteString(" ")

	if h.opts.includeTimestamp && !r.Time.IsZero() {
		timestamp := r.Time.Format(h.opts.timestampFormat)
		b.WriteString(h.colorize(timestamp, h.opts.colors, FgGray))
		b.WriteString(" ")
	}

	if h.opts.includeSource && r.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		source := fmt.Sprintf("%s:%d", frame.File, frame.Line)
		b.WriteString(h.colorize(source, h.opts.colors, FgGray))
		b.WriteString(" ")
	}

	b.WriteString(h.colorize(r.Message, h.opts.colors, ""))

	attrs := h.collectAttrs(r)

	levelColor := h.opts.levelColors[r.Level]
	if levelColor == "" {
		levelColor = FgWhite
	}

	if len(attrs) > 0 {
		b.WriteString(h.formatAttrs(attrs, levelColor))
	}

	b.WriteString("\n")
	_, err := h.writer.Write([]byte(b.String()))
	return err
}

func (h *PrettyslogHandler) colorize(text string, enable bool, colorCode string) string {
	if !enable || colorCode == "" {
		return text
	}
	resetCode := ColorReset
	return fmt.Sprintf("%s%s%s", colorCode, text, resetCode)
}

func (h *PrettyslogHandler) formatLevel(level slog.Level) string {
	var levelStr string
	switch level {
	case slog.LevelDebug:
		levelStr = "DBG"
	case slog.LevelInfo:
		levelStr = "INF"
	case slog.LevelWarn:
		levelStr = "WAR"
	case slog.LevelError:
		levelStr = "ERR"
	default:
		levelStr = "UNK"
	}
	colorCode := h.opts.levelColors[level]
	if colorCode == "" {
		colorCode = FgWhite
	}
	return h.colorize(levelStr, h.opts.colors, colorCode)
}

func (h *PrettyslogHandler) collectAttrs(r slog.Record) []slog.Attr {
	var attrs []slog.Attr
	attrs = append(attrs, h.attrs...)

	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	return attrs
}

func (h *PrettyslogHandler) formatAttrs(attrs []slog.Attr, levelColor string) string {
	var b strings.Builder
	for _, attr := range attrs {
		b.WriteString(h.colorize(";", h.opts.colors, FgGray))
		b.WriteString(" ")

		h.formatAttr(&b, attr, levelColor, "")
	}
	return b.String()
}

func (h *PrettyslogHandler) formatAttr(b *strings.Builder, attr slog.Attr, levelColor string, prefix string) {
	if attr.Value.Kind() == slog.KindGroup {
		groupName := attr.Key
		groupAttrs := attr.Value.Group()
		newPrefix := groupName
		if prefix != "" {
			newPrefix = prefix + "." + groupName
		}
		for _, a := range groupAttrs {
			b.WriteString(h.colorize(";", h.opts.colors, FgGray))
			b.WriteString(" ")

			h.formatAttr(b, a, levelColor, newPrefix)
		}
	} else {
		key := attr.Key
		if prefix != "" {
			key = prefix + "." + key
		}
		keyColored := h.colorize(key, h.opts.colors, levelColor)
		b.WriteString(keyColored)

		b.WriteString(h.colorize("=", h.opts.colors, FgGray))

		value := h.colorize(fmt.Sprintf("%v", attr.Value.Any()), h.opts.colors, "")
		b.WriteString(value)
	}
}

func (h *PrettyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := *h
	h2.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &h2
}

func (h *PrettyslogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := *h
	h2.groups = append(append([]string{}, h.groups...), name)
	return &h2
}
