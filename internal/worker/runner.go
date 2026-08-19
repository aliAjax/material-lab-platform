package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Job struct {
	ID          string
	Kind        string
	Payload     map[string]string
	Attempts    int
	MaxAttempts int
	AvailableAt time.Time
	LeaseOwner  string
	LeaseUntil  time.Time
	CompletedAt time.Time
	LastError   string
}

type Queue interface {
	Lease(context.Context, string, time.Duration) (Job, error)
	Complete(context.Context, string, string) error
	Fail(context.Context, string, string, error, time.Time) error
}

var ErrNoJobs = errors.New("no jobs available")

type MemoryQueue struct {
	mu   sync.Mutex
	jobs map[string]Job
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{jobs: make(map[string]Job)}
}

func (q *MemoryQueue) Enqueue(job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, exists := q.jobs[job.ID]; exists {
		return fmt.Errorf("job already exists")
	}
	if job.MaxAttempts == 0 {
		job.MaxAttempts = 5
	}
	q.jobs[job.ID] = job
	return nil
}

func (q *MemoryQueue) Lease(ctx context.Context, owner string, duration time.Duration) (Job, error) {
	if err := ctx.Err(); err != nil {
		return Job{}, err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now().UTC()
	for id, job := range q.jobs {
		if !job.CompletedAt.IsZero() || job.Attempts >= job.MaxAttempts {
			continue
		}
		if now.Before(job.AvailableAt) || now.Before(job.LeaseUntil) {
			continue
		}
		job.LeaseOwner = owner
		job.LeaseUntil = now.Add(duration)
		job.Attempts++
		q.jobs[id] = job
		return job, nil
	}
	return Job{}, ErrNoJobs
}

func (q *MemoryQueue) Complete(ctx context.Context, id, owner string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	job, exists := q.jobs[id]
	if !exists {
		return fmt.Errorf("job not found")
	}
	if job.LeaseOwner != owner || time.Now().UTC().After(job.LeaseUntil) {
		return fmt.Errorf("job lease lost")
	}
	job.CompletedAt = time.Now().UTC()
	job.LeaseUntil = time.Time{}
	q.jobs[id] = job
	return nil
}

func (q *MemoryQueue) Fail(ctx context.Context, id, owner string, cause error, retryAt time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	job, exists := q.jobs[id]
	if !exists {
		return fmt.Errorf("job not found")
	}
	if job.LeaseOwner != owner {
		return fmt.Errorf("job lease owner mismatch")
	}
	job.LastError = cause.Error()
	job.AvailableAt = retryAt
	job.LeaseOwner = ""
	job.LeaseUntil = time.Time{}
	q.jobs[id] = job
	return nil
}

type Handler func(context.Context, Job) error

type Runner struct {
	queue         Queue
	owner         string
	leaseDuration time.Duration
	handlers      map[string]Handler
	pollInterval  time.Duration
}

func NewRunner(queue Queue, owner string, leaseDuration time.Duration) *Runner {
	return &Runner{
		queue:         queue,
		owner:         owner,
		leaseDuration: leaseDuration,
		handlers:      make(map[string]Handler),
		pollInterval:  time.Second,
	}
}

func (r *Runner) Register(kind string, handler Handler) {
	r.handlers[kind] = handler
}

func (r *Runner) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()
	for {
		if err := r.runOne(ctx); err != nil && !errors.Is(err, ErrNoJobs) {
			slog.Error("job loop failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (r *Runner) runOne(ctx context.Context) error {
	job, err := r.queue.Lease(ctx, r.owner, r.leaseDuration)
	if err != nil {
		return err
	}
	handler, exists := r.handlers[job.Kind]
	if !exists {
		err = fmt.Errorf("unknown job kind %s", job.Kind)
	} else {
		err = handler(ctx, job)
	}
	if err == nil {
		return r.queue.Complete(ctx, job.ID, r.owner)
	}
	delay := time.Duration(job.Attempts*job.Attempts) * time.Second
	return r.queue.Fail(ctx, job.ID, r.owner, err, time.Now().UTC().Add(delay))
}

func NotificationHandler(logger *slog.Logger) Handler {
	return func(ctx context.Context, job Job) error {
		logger.InfoContext(ctx, "notification dispatched", "job_id", job.ID, "user_id", job.Payload["userId"])
		return nil
	}
}

func CertificateHandler(root string, logger *slog.Logger) Handler {
	return func(ctx context.Context, job Job) error {
		if err := os.MkdirAll(root, 0o750); err != nil {
			return err
		}
		name := filepath.Base(job.Payload["certificateId"]) + ".html"
		if name == ".html" {
			return fmt.Errorf("certificate id required")
		}
		path := filepath.Join(root, name)
		content := []byte("<!doctype html><html lang=\"zh-CN\"><meta charset=\"utf-8\"><title>检验证书</title><body><h1>检验证书</h1><p>编号：" + job.Payload["number"] + "</p><p>摘要：" + job.Payload["digest"] + "</p></body></html>")
		if err := os.WriteFile(path, content, 0o640); err != nil {
			return err
		}
		logger.InfoContext(ctx, "certificate generated", "path", path)
		return nil
	}
}
