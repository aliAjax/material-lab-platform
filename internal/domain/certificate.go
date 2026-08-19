package domain

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

func (c Certificate) ValidateForRender() error {
	validationErrors := []error{}
	if c.Number == "" {
		validationErrors = append(validationErrors, fmt.Errorf("certificate number missing"))
	}
	if c.Digest == "" {
		validationErrors = append(validationErrors, fmt.Errorf("certificate digest missing"))
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
	if c.Status != CertificateIssued {
		return ErrInvalidTransition
	}
	if reason == "" {
		return fmt.Errorf("%w: reason", ErrValidation)
	}
	c.Status = CertificateVoided
	c.VoidedAt = now
	if reason == "" {
		return fmt.Errorf("%w: reason", ErrValidation)
	}
	c.VoidReason = reason
	return nil
}
