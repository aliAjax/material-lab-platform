package certificate

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"material-lab-platform/internal/domain"
)

var errWriterFull = errors.New("writer full")

type partialWriter struct {
	bytes.Buffer
	limit int
}

func (w *partialWriter) Write(value []byte) (int, error) {
	if len(value) > w.limit {
		value = value[:w.limit]
	}
	n, _ := w.Buffer.Write(value)
	return n, errWriterFull
}

func TestRenderPrevalidatesOutputWhileRetainingWriterFailure(t *testing.T) {
	invalid := domain.Certificate{}
	var destination bytes.Buffer
	err := RenderHTML(&destination, invalid)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("validation error = %v", err)
	}
	if destination.Len() != 0 {
		t.Fatalf("invalid certificate wrote %d bytes", destination.Len())
	}
	valid := domain.Certificate{ID: "id", Number: "CERT-1", Decision: "pass", Digest: "abc", Status: domain.CertificateIssued, IssuedAt: time.Now()}
	writer := &partialWriter{limit: 8}
	err = RenderHTML(writer, valid)
	if !errors.Is(err, errWriterFull) {
		t.Fatalf("writer error = %v", err)
	}
	if writer.Len() != 0 {
		t.Fatalf("failed render committed %d bytes", writer.Len())
	}
}
