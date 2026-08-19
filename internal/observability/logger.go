package observability

import (
	"context"
	"log/slog"
	"strings"
)

type RedactingHandler struct {
	next slog.Handler
}

func NewRedactingHandler(next slog.Handler) *RedactingHandler {
	return &RedactingHandler{next: next}
}

func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *RedactingHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attribute slog.Attr) bool {
		clean.AddAttrs(redact(attribute))
		return true
	})
	return h.next.Handle(ctx, clean)
}

func (h *RedactingHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, len(attributes))
	for i, attribute := range attributes {
		clean[i] = redact(attribute)
	}
	return &RedactingHandler{next: h.next.WithAttrs(clean)}
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{next: h.next.WithGroup(name)}
}

func redact(attribute slog.Attr) slog.Attr {
	lower := strings.ToLower(attribute.Key)
	if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") {
		return slog.String(attribute.Key, "[REDACTED]")
	}
	return attribute
}
