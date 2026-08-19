package domain

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskReview     TaskStatus = "review"
	TaskApproved   TaskStatus = "approved"
	TaskRejected   TaskStatus = "rejected"
)

type Task struct {
	ID            string     `json:"id"`
	RequestID     string     `json:"requestId"`
	SubsampleID   string     `json:"subsampleId"`
	MethodID      string     `json:"methodId"`
	MethodVersion int        `json:"methodVersion"`
	AssigneeID    string     `json:"assigneeId"`
	ExecutorID    string     `json:"executorId,omitempty"`
	Status        TaskStatus `json:"status"`
	CurrentRound  int        `json:"currentRound"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
type TaskRound struct {
	ID        string                     `json:"id"`
	TaskID    string                     `json:"taskId"`
	Number    int                        `json:"number"`
	Readings  map[string]decimal.Decimal `json:"readings"`
	Result    decimal.Decimal            `json:"result"`
	Decision  string                     `json:"decision"`
	Reason    string                     `json:"reason,omitempty"`
	CreatedBy string                     `json:"createdBy"`
	CreatedAt time.Time                  `json:"createdAt"`
}
type RetestRequest struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"taskId"`
	Round          int       `json:"round"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	DecidedBy      string    `json:"decidedBy,omitempty"`
	DecisionReason string    `json:"decisionReason,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (t *Task) Start(actor string, hasCustody bool, now time.Time) error {
	if t.Status != TaskPending {
		return ErrInvalidTransition
	}
	if actor != t.AssigneeID || !hasCustody {
		return fmt.Errorf("%w: custody confirmation required", ErrForbidden)
	}
	t.ExecutorID = actor
	t.Status = TaskInProgress
	t.CurrentRound = 1
	t.UpdatedAt = now
	return nil
}
func (t *Task) SubmitReview(actor string, round TaskRound, now time.Time) error {
	if t.Status != TaskInProgress || actor != t.ExecutorID {
		return ErrInvalidTransition
	}
	if round.Number != t.CurrentRound {
		return fmt.Errorf("%w: round", ErrValidation)
	}
	t.Status = TaskReview
	t.UpdatedAt = now
	return nil
}
func (t *Task) Approve(actor string, now time.Time) error {
	if t.Status != TaskReview || actor == t.ExecutorID {
		return ErrForbidden
	}
	t.Status = TaskApproved
	t.UpdatedAt = now
	return nil
}
func (t *Task) Reject(actor string, now time.Time) error {
	if t.Status != TaskReview || actor == t.ExecutorID {
		return ErrForbidden
	}
	t.Status = TaskRejected
	t.UpdatedAt = now
	return nil
}
func (t *Task) StartRetest(now time.Time) error {
	if t.Status != TaskRejected {
		return ErrInvalidTransition
	}
	t.CurrentRound++
	t.Status = TaskInProgress
	t.UpdatedAt = now
	return nil
}
