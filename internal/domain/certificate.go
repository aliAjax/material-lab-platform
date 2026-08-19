package domain

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

func (c Certificate) ValidateForRender() error {
	validationErrors := make([]error, 0)
	for name, value := range map[string]string{"id": c.ID, "number": c.Number, "decision": c.Decision, "digest": c.Digest} {
		if value == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%w: certificate %s", ErrValidation, name))
		}
	}
	if c.Status != CertificateIssued && c.Status != CertificateVoided {
		validationErrors = append(validationErrors, fmt.Errorf("%w: certificate status", ErrValidation))
	}
	if c.IssuedAt.IsZero() {
		validationErrors = append(validationErrors, fmt.Errorf("%w: certificate issuedAt", ErrValidation))
	}
	return errors.Join(validationErrors...)
}

type CertificateStatus string

const (
	CertificateIssued CertificateStatus = "issued"
	CertificateVoided CertificateStatus = "voided"
)

type Certificate struct {
	ID               string                     `json:"id"`
	Number           string                     `json:"number"`
	RequestID        string                     `json:"requestId"`
	SampleID         string                     `json:"sampleId"`
	MethodVersions   map[string]int             `json:"methodVersions"`
	Results          map[string]decimal.Decimal `json:"results"`
	Decision         string                     `json:"decision"`
	ExecutorIDs      []string                   `json:"executorIds"`
	ReviewerIDs      []string                   `json:"reviewerIds"`
	Status           CertificateStatus          `json:"status"`
	VerificationCode string                     `json:"-"`
	Digest           string                     `json:"digest"`
	SupersedesID     string                     `json:"supersedesId,omitempty"`
	IssuedBy         string                     `json:"issuedBy"`
	IssuedAt         time.Time                  `json:"issuedAt"`
	VoidedAt         time.Time                  `json:"voidedAt,omitempty"`
	VoidReason       string                     `json:"voidReason,omitempty"`
}

func (c *Certificate) Void(actor, reason string, now time.Time) error {
	if c == nil {
		return ErrNotFound
	}
	if c.Status != CertificateIssued {
		return ErrInvalidTransition
	}
	if actor == "" || reason == "" || now.IsZero() {
		return fmt.Errorf("%w: reason", ErrValidation)
	}
	next := *c
	next.Status = CertificateVoided
	next.VoidedAt = now.UTC()
	next.VoidReason = reason
	*c = next
	return nil
}
