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
	attribute.Value = attribute.Value.Resolve()
	if attribute.Value.Kind() == slog.KindGroup {
		members := attribute.Value.Group()
		clean := make([]slog.Attr, len(members))
		for i, member := range members {
			clean[i] = redact(member)
		}
		return slog.Group(attribute.Key, attrsToAny(clean)...)
	}
	lower := strings.ToLower(attribute.Key)
	if sensitiveKey(lower) {
		return slog.String(attribute.Key, "[REDACTED]")
	}
	return attribute
}

func sensitiveKey(lower string) bool {
	for _, fragment := range []string{"password", "token", "secret", "authorization", "cookie", "content"} {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

func attrsToAny(attributes []slog.Attr) []any {
	values := make([]any, len(attributes))
	for i, attribute := range attributes {
		values[i] = attribute
	}
	return values
}
