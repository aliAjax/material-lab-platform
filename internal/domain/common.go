package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	ErrForbidden         = errors.New("operation forbidden")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrValidation        = errors.New("validation failed")
)

type Role string

const (
	RoleRegistrar Role = "registrar"
	RoleTester    Role = "tester"
	RoleReviewer  Role = "reviewer"
	RoleManager   Role = "manager"
)

func (r Role) Valid() bool {
	switch r {
	case RoleRegistrar, RoleTester, RoleReviewer, RoleManager:
		return true
	default:
		return false
	}
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NewID() string { return uuid.NewString() }

func RandomToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func SecureDigestEqual(value, expected string) bool {
	return Digest(value) == expected
}

func SplitOpaqueToken(raw string) (string, string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return "", "", ErrValidation
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", ErrValidation
	}
	return parts[0], parts[1], nil
}

func Require(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s is required", ErrValidation, field)
	}
	return nil
}

type ChangeSummary struct {
	Before map[string]string `json:"before,omitempty"`
	After  map[string]string `json:"after,omitempty"`
}

type AuditEvent struct {
	ID        string        `json:"id"`
	ActorID   string        `json:"actorId"`
	Action    string        `json:"action"`
	Object    string        `json:"object"`
	ObjectID  string        `json:"objectId"`
	RequestID string        `json:"requestId"`
	Change    ChangeSummary `json:"change"`
	CreatedAt time.Time     `json:"createdAt"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	ObjectID  string    `json:"objectId"`
	ReadAt    time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

func StableSort[T any](items []T, less func(a, b T) bool) {
	sort.SliceStable(items, func(i, j int) bool { return less(items[i], items[j]) })
}

func RedactChange(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "content") {
			result[key] = "[REDACTED]"
			continue
		}
		result[key] = value
	}
	return result
}
