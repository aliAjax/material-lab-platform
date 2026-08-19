package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestCertificateVoidFailureDoesNotMutateState(t *testing.T) {
	certificate := Certificate{Status: CertificateIssued}
	before := certificate
	err := certificate.Void("operator", "", time.Now())
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("void error = %v", err)
	}
	if !reflect.DeepEqual(certificate, before) {
		t.Fatalf("certificate mutated after failure: %#v", certificate)
	}
	var missing *Certificate
	if !errors.Is(missing.Void("operator", "reason", time.Now()), ErrNotFound) {
		t.Fatal("nil certificate did not return not found")
	}
}
