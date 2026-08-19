package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type RequestStatus string

const (
	RequestDraft     RequestStatus = "draft"
	RequestSubmitted RequestStatus = "submitted"
	RequestCompleted RequestStatus = "completed"
)

type ClientSource struct {
	Organization string `json:"organization"`
	ContactName  string `json:"contactName,omitempty"`
	ContactPhone string `json:"contactPhone,omitempty"`
}

type InspectionRequest struct {
	ID                string        `json:"id"`
	Reference         string        `json:"reference"`
	Client            ClientSource  `json:"client"`
	MaterialGrade     string        `json:"materialGrade"`
	BatchNumber       string        `json:"batchNumber"`
	SampleDescription string        `json:"sampleDescription"`
	Requirements      string        `json:"requirements"`
	ReceivedAt        time.Time     `json:"receivedAt"`
	Status            RequestStatus `json:"status"`
	SampleID          string        `json:"sampleId,omitempty"`
	Revision          int           `json:"revision"`
	CreatedBy         string        `json:"createdBy"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

func NewInspectionRequest(actor string, now time.Time) InspectionRequest {
	return InspectionRequest{ID: NewID(), Status: RequestDraft, Revision: 1, CreatedBy: actor, CreatedAt: now, UpdatedAt: now}
}

func (r *InspectionRequest) ValidateForSubmit() error {
	checks := []struct{ value, field string }{
		{r.Client.Organization, "client.organization"},
		{r.MaterialGrade, "materialGrade"},
		{r.BatchNumber, "batchNumber"},
		{r.SampleDescription, "sampleDescription"},
		{r.Requirements, "requirements"},
	}
	for _, check := range checks {
		if err := Require(check.value, check.field); err != nil {
			return err
		}
	}
	if r.ReceivedAt.IsZero() {
		return fmt.Errorf("%w: receivedAt is required", ErrValidation)
	}
	return nil
}

func (r *InspectionRequest) Submit(number, sampleID string, now time.Time) error {
	if r.Status != RequestDraft {
		return ErrInvalidTransition
	}
	if err := r.ValidateForSubmit(); err != nil {
		return err
	}
	r.Reference, r.SampleID, r.Status, r.UpdatedAt = number, sampleID, RequestSubmitted, now
	return nil
}

type SampleStatus string

const (
	SampleRegistered SampleStatus = "registered"
	SampleSplit      SampleStatus = "split"
	SampleInTesting  SampleStatus = "in_testing"
	SampleSealed     SampleStatus = "sealed"
	SampleDisposed   SampleStatus = "disposed"
)

type Sample struct {
	ID          string          `json:"id"`
	RequestID   string          `json:"requestId"`
	Number      string          `json:"number"`
	Description string          `json:"description"`
	TotalQty    decimal.Decimal `json:"totalQuantity"`
	Unit        string          `json:"unit"`
	Status      SampleStatus    `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type Subsample struct {
	ID            string          `json:"id"`
	SampleID      string          `json:"sampleId"`
	ParentID      string          `json:"parentId,omitempty"`
	Label         string          `json:"label"`
	Purpose       string          `json:"purpose"`
	Quantity      decimal.Decimal `json:"quantity"`
	Unit          string          `json:"unit"`
	LocationID    string          `json:"locationId"`
	Status        SampleStatus    `json:"status"`
	CurrentHolder string          `json:"currentHolder,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type SplitInput struct {
	Label      string          `json:"label"`
	Purpose    string          `json:"purpose"`
	Quantity   decimal.Decimal `json:"quantity"`
	Unit       string          `json:"unit"`
	LocationID string          `json:"locationId"`
}

func ValidateSplit(parent Sample, existing []Subsample, additions []SplitInput, loss decimal.Decimal, lossReason string) error {
	if len(additions) == 0 {
		return fmt.Errorf("%w: at least one subsample required", ErrValidation)
	}
	if loss.IsNegative() {
		return fmt.Errorf("%w: loss cannot be negative", ErrValidation)
	}
	if loss.GreaterThan(decimal.Zero) && strings.TrimSpace(lossReason) == "" {
		return fmt.Errorf("%w: loss reason required", ErrValidation)
	}
	used := loss
	labels := map[string]bool{}
	for _, item := range existing {
		used = used.Add(item.Quantity)
		labels[item.Label] = true
	}
	for _, item := range additions {
		if strings.TrimSpace(item.Label) == "" || strings.TrimSpace(item.Purpose) == "" {
			return fmt.Errorf("%w: label and purpose required", ErrValidation)
		}
		if item.Unit != parent.Unit {
			return fmt.Errorf("%w: unit mismatch", ErrValidation)
		}
		if !item.Quantity.GreaterThan(decimal.Zero) {
			return fmt.Errorf("%w: quantity must be positive", ErrValidation)
		}
		if labels[item.Label] {
			return fmt.Errorf("%w: duplicate label %s", ErrConflict, item.Label)
		}
		labels[item.Label] = true
		used = used.Add(item.Quantity)
	}
	if used.GreaterThan(parent.TotalQty) {
		return fmt.Errorf("%w: allocated %s exceeds total %s", ErrValidation, used, parent.TotalQty)
	}
	return nil
}

type Location struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
