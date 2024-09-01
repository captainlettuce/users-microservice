package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
)

type contextKey string
type HandlerType string

const (
	slogContextKey contextKey = "slogFields"

	HandlerTypeJSON HandlerType = "json"
	HandlerTypeText HandlerType = "text"
)

type LoggerSettings struct {
	level      slog.Level
	skipSource bool
	output     io.Writer
	handler    HandlerType
}

type Option func(*LoggerSettings)

// WithSkipSource removes source-code reference of the caller from log output
func WithSkipSource() Option {
	return func(settings *LoggerSettings) {
		settings.skipSource = true
	}
}

// WithOutput specifies a custom writer as output for the logger
func WithOutput(output io.Writer) Option {
	return func(settings *LoggerSettings) {
		settings.output = output
	}
}

// WithLevel sets the minimum severity level of messages to log
func WithLevel(level slog.Level) Option {
	return func(settings *LoggerSettings) {
		settings.level = level
	}
}

// WithHandler sets the handler to use
// implemented options are "json" and "text"
func WithHandler(handler HandlerType) Option {
	return func(settings *LoggerSettings) {
		settings.handler = handler
	}
}

func New(serviceName string, options ...Option) (*slog.Logger, error) {
	settings := &LoggerSettings{
		level:   slog.LevelInfo,
		output:  os.Stdout,
		handler: HandlerTypeJSON,
	}

	for _, option := range options {
		option(settings)
	}

	handlerOptions := slog.HandlerOptions{
		AddSource:   !settings.skipSource,
		Level:       settings.level,
		ReplaceAttr: nil,
	}

	var handler slog.Handler
	switch settings.handler {
	case HandlerTypeJSON:
		handler = slog.NewJSONHandler(settings.output, &handlerOptions)
	case HandlerTypeText:
		handler = slog.NewTextHandler(settings.output, &handlerOptions)
	default:
		return nil, errors.New("unknown handler")
	}

	logger := slog.New(&ContextHandler{handler}).With(
		slog.String("service", serviceName),
	)

	bi, ok := debug.ReadBuildInfo()
	if ok {
		logger = logger.With(
			slog.String("go_version", bi.GoVersion),
			slog.String("version", bi.Main.Version),
		)
	}

	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" || s.Key == "vcs.modified" {
			logger = logger.With(
				slog.String(s.Key, s.Value))
		}
	}

	return logger, nil
}

// ContextHandler is a wrapper that injects slog.Attrs from Context
type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogContextKey).([]slog.Attr); ok {
		for _, attr := range attrs {
			r.AddAttrs(attr)
		}
	}

	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{h.Handler.WithGroup(name)}
}

// LogToContext adds a slog.Attr to the context which will then be logged for all subsequent calls to the logger
func LogToContext(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var existing []slog.Attr
	if a, ok := ctx.Value(slogContextKey).([]slog.Attr); ok {
		existing = a
	}

	existing = append(existing, attrs...)

	return context.WithValue(ctx, slogContextKey, existing)
}
