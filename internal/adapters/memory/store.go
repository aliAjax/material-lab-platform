package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"material-lab-platform/internal/domain"
	"material-lab-platform/internal/ports"
)

type Store struct {
	mu           sync.RWMutex
	sequence     int64
	requests     map[string]domain.InspectionRequest
	samples      map[string]domain.Sample
	subsamples   map[string]domain.Subsample
	custodies    map[string]domain.CustodyTransfer
	methods      map[string]domain.Method
	tasks        map[string]domain.Task
	rounds       map[string]domain.TaskRound
	retests      map[string]domain.RetestRequest
	certificates map[string]domain.Certificate
	audits       []domain.AuditEvent
}

func New() *Store {
	return &Store{requests: map[string]domain.InspectionRequest{}, samples: map[string]domain.Sample{}, subsamples: map[string]domain.Subsample{}, custodies: map[string]domain.CustodyTransfer{}, methods: map[string]domain.Method{}, tasks: map[string]domain.Task{}, rounds: map[string]domain.TaskRound{}, retests: map[string]domain.RetestRequest{}, certificates: map[string]domain.Certificate{}, audits: []domain.AuditEvent{}}
}

func (s *Store) Transaction(ctx context.Context, fn func(ports.Repository) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(&locked{s})
}

type locked struct{ s *Store }

func (s *Store) SaveRequest(x domain.InspectionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveRequest(x)
}
func (l *locked) SaveRequest(x domain.InspectionRequest) error { l.s.requests[x.ID] = x; return nil }
func (s *Store) Request(id string) (domain.InspectionRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Request(id)
}
func (l *locked) Request(id string) (domain.InspectionRequest, error) {
	x, ok := l.s.requests[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) Requests() []domain.InspectionRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Requests()
}
func (l *locked) Requests() []domain.InspectionRequest {
	out := make([]domain.InspectionRequest, 0, len(l.s.requests))
	for _, x := range l.s.requests {
		out = append(out, x)
	}
	domain.StableSort(out, func(a, b domain.InspectionRequest) bool { return a.CreatedAt.After(b.CreatedAt) })
	return out
}
func (s *Store) NextReference() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).NextReference()
}
func (l *locked) NextReference() string {
	l.s.sequence++
	return fmt.Sprintf("SMP-%s-%05d", time.Now().UTC().Format("20060102"), l.s.sequence)
}
func (s *Store) SaveSample(x domain.Sample) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveSample(x)
}
func (l *locked) SaveSample(x domain.Sample) error { l.s.samples[x.ID] = x; return nil }
func (s *Store) Sample(id string) (domain.Sample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Sample(id)
}
func (l *locked) Sample(id string) (domain.Sample, error) {
	x, ok := l.s.samples[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) Samples() []domain.Sample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Samples()
}
func (l *locked) Samples() []domain.Sample {
	out := []domain.Sample{}
	for _, x := range l.s.samples {
		out = append(out, x)
	}
	return out
}
func (s *Store) SaveSubsample(x domain.Subsample) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveSubsample(x)
}
func (l *locked) SaveSubsample(x domain.Subsample) error { l.s.subsamples[x.ID] = x; return nil }
func (s *Store) Subsamples(sample string) []domain.Subsample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Subsamples(sample)
}
func (l *locked) Subsamples(sample string) []domain.Subsample {
	out := []domain.Subsample{}
	for _, x := range l.s.subsamples {
		if x.SampleID == sample {
			out = append(out, x)
		}
	}
	return out
}
func (s *Store) Subsample(id string) (domain.Subsample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Subsample(id)
}
func (l *locked) Subsample(id string) (domain.Subsample, error) {
	x, ok := l.s.subsamples[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) SaveCustody(x domain.CustodyTransfer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveCustody(x)
}
func (l *locked) SaveCustody(x domain.CustodyTransfer) error { l.s.custodies[x.ID] = x; return nil }
func (s *Store) Custody(id string) (domain.CustodyTransfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Custody(id)
}
func (l *locked) Custody(id string) (domain.CustodyTransfer, error) {
	x, ok := l.s.custodies[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) Custodies() []domain.CustodyTransfer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Custodies()
}
func (l *locked) Custodies() []domain.CustodyTransfer {
	out := []domain.CustodyTransfer{}
	for _, x := range l.s.custodies {
		out = append(out, x)
	}
	return out
}
func (s *Store) SaveMethod(x domain.Method) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveMethod(x)
}
func (l *locked) SaveMethod(x domain.Method) error { l.s.methods[x.ID] = x; return nil }
func (s *Store) Method(id string) (domain.Method, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Method(id)
}
func (l *locked) Method(id string) (domain.Method, error) {
	x, ok := l.s.methods[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) Methods() []domain.Method {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Methods()
}
func (l *locked) Methods() []domain.Method {
	out := []domain.Method{}
	for _, x := range l.s.methods {
		out = append(out, x)
	}
	return out
}
func (s *Store) SaveTask(x domain.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveTask(x)
}
func (l *locked) SaveTask(x domain.Task) error { l.s.tasks[x.ID] = x; return nil }
func (s *Store) Task(id string) (domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Task(id)
}
func (l *locked) Task(id string) (domain.Task, error) {
	x, ok := l.s.tasks[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) Tasks() []domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Tasks()
}
func (l *locked) Tasks() []domain.Task {
	out := []domain.Task{}
	for _, x := range l.s.tasks {
		out = append(out, x)
	}
	return out
}
func (s *Store) SaveRound(x domain.TaskRound) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveRound(x)
}
func (l *locked) SaveRound(x domain.TaskRound) error { l.s.rounds[x.ID] = x; return nil }
func (s *Store) Rounds(task string) []domain.TaskRound {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Rounds(task)
}
func (l *locked) Rounds(task string) []domain.TaskRound {
	out := []domain.TaskRound{}
	for _, x := range l.s.rounds {
		if x.TaskID == task {
			out = append(out, x)
		}
	}
	return out
}
func (s *Store) SaveRetest(x domain.RetestRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveRetest(x)
}
func (l *locked) SaveRetest(x domain.RetestRequest) error { l.s.retests[x.ID] = x; return nil }
func (s *Store) Retests(task string) []domain.RetestRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Retests(task)
}
func (l *locked) Retests(task string) []domain.RetestRequest {
	out := []domain.RetestRequest{}
	for _, x := range l.s.retests {
		if x.TaskID == task {
			out = append(out, x)
		}
	}
	return out
}
func (s *Store) SaveCertificate(x domain.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).SaveCertificate(x)
}
func (l *locked) SaveCertificate(x domain.Certificate) error { l.s.certificates[x.ID] = x; return nil }
func (s *Store) Certificate(id string) (domain.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Certificate(id)
}
func (l *locked) Certificate(id string) (domain.Certificate, error) {
	x, ok := l.s.certificates[id]
	if !ok {
		return x, domain.ErrNotFound
	}
	return x, nil
}
func (s *Store) CertificateByCode(code string) (domain.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).CertificateByCode(code)
}
func (l *locked) CertificateByCode(code string) (domain.Certificate, error) {
	for _, x := range l.s.certificates {
		if domain.Digest(code) == domain.Digest(x.VerificationCode) {
			return x, nil
		}
	}
	return domain.Certificate{}, domain.ErrNotFound
}
func (s *Store) Certificates() []domain.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Certificates()
}
func (l *locked) Certificates() []domain.Certificate {
	out := []domain.Certificate{}
	for _, x := range l.s.certificates {
		out = append(out, x)
	}
	return out
}
func (s *Store) AppendAudit(x domain.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return (&locked{s}).AppendAudit(x)
}
func (l *locked) AppendAudit(x domain.AuditEvent) error {
	l.s.audits = append(l.s.audits, x)
	return nil
}
func (s *Store) Audits() []domain.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return (&locked{s}).Audits()
}
func (l *locked) Audits() []domain.AuditEvent { return append([]domain.AuditEvent(nil), l.s.audits...) }
