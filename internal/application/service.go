package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"material-lab-platform/internal/domain"
	"material-lab-platform/internal/ports"
)

type Service struct {
	store ports.Store
	clock func() time.Time
}

func New(store ports.Store) *Service {
	return &Service{store: store, clock: func() time.Time { return time.Now().UTC() }}
}
func (s *Service) audit(repo ports.Repository, actor, action, kind, id, requestID string, after map[string]string) error {
	return repo.AppendAudit(domain.AuditEvent{ID: domain.NewID(), ActorID: actor, Action: action, Object: kind, ObjectID: id, RequestID: requestID, Change: domain.ChangeSummary{After: domain.RedactChange(after)}, CreatedAt: s.clock()})
}

type RequestInput struct {
	Organization      string          `json:"organization"`
	ContactName       string          `json:"contactName"`
	ContactPhone      string          `json:"contactPhone"`
	MaterialGrade     string          `json:"materialGrade"`
	BatchNumber       string          `json:"batchNumber"`
	SampleDescription string          `json:"sampleDescription"`
	Requirements      string          `json:"requirements"`
	ReceivedAt        time.Time       `json:"receivedAt"`
	TotalQuantity     decimal.Decimal `json:"totalQuantity"`
	Unit              string          `json:"unit"`
}

func (s *Service) CreateRequest(ctx context.Context, actor, requestID string, in RequestInput) (domain.InspectionRequest, error) {
	r := domain.NewInspectionRequest(actor, s.clock())
	r.Client = domain.ClientSource{Organization: in.Organization, ContactName: in.ContactName, ContactPhone: in.ContactPhone}
	r.MaterialGrade = in.MaterialGrade
	r.BatchNumber = in.BatchNumber
	r.SampleDescription = in.SampleDescription
	r.Requirements = in.Requirements
	r.ReceivedAt = in.ReceivedAt
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		if err := repo.SaveRequest(r); err != nil {
			return err
		}
		return s.audit(repo, actor, "create", "commission", r.ID, requestID, map[string]string{"status": string(r.Status)})
	})
	return r, err
}
func (s *Service) UpdateRequest(ctx context.Context, actor, id, requestID string, in RequestInput) (domain.InspectionRequest, error) {
	var r domain.InspectionRequest
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		r, err = repo.Request(id)
		if err != nil {
			return err
		}
		if r.Status != domain.RequestDraft {
			return domain.ErrInvalidTransition
		}
		r.Client = domain.ClientSource{Organization: in.Organization, ContactName: in.ContactName, ContactPhone: in.ContactPhone}
		r.MaterialGrade = in.MaterialGrade
		r.BatchNumber = in.BatchNumber
		r.SampleDescription = in.SampleDescription
		r.Requirements = in.Requirements
		r.ReceivedAt = in.ReceivedAt
		r.Revision++
		r.UpdatedAt = s.clock()
		if err = repo.SaveRequest(r); err != nil {
			return err
		}
		return s.audit(repo, actor, "update", "commission", id, requestID, map[string]string{"revision": strconv.Itoa(r.Revision)})
	})
	return r, err
}
func (s *Service) SubmitRequest(ctx context.Context, actor, id, requestID string, total decimal.Decimal, unit string) (domain.InspectionRequest, domain.Sample, error) {
	var r domain.InspectionRequest
	var sample domain.Sample
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		r, err = repo.Request(id)
		if err != nil {
			return err
		}
		if !total.GreaterThan(decimal.Zero) || unit == "" {
			return fmt.Errorf("%w: quantity and unit", domain.ErrValidation)
		}
		number := repo.NextReference()
		sample = domain.Sample{ID: domain.NewID(), RequestID: r.ID, Number: number, Description: r.SampleDescription, TotalQty: total, Unit: unit, Status: domain.SampleRegistered, CreatedAt: s.clock()}
		if err = r.Submit(number, sample.ID, s.clock()); err != nil {
			return err
		}
		if err = repo.SaveSample(sample); err != nil {
			return err
		}
		if err = repo.SaveRequest(r); err != nil {
			return err
		}
		return s.audit(repo, actor, "submit", "commission", id, requestID, map[string]string{"sampleNumber": number})
	})
	return r, sample, err
}
func (s *Service) SplitSample(ctx context.Context, actor, id, requestID string, parts []domain.SplitInput, loss decimal.Decimal, reason string) ([]domain.Subsample, error) {
	var created []domain.Subsample
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		sample, err := repo.Sample(id)
		if err != nil {
			return err
		}
		existing := repo.Subsamples(id)
		if err = domain.ValidateSplit(sample, existing, parts, loss, reason); err != nil {
			return err
		}
		for _, in := range parts {
			x := domain.Subsample{ID: domain.NewID(), SampleID: id, Label: in.Label, Purpose: in.Purpose, Quantity: in.Quantity, Unit: in.Unit, LocationID: in.LocationID, Status: domain.SampleRegistered, CreatedAt: s.clock()}
			if err = repo.SaveSubsample(x); err != nil {
				return err
			}
			created = append(created, x)
		}
		sample.Status = domain.SampleSplit
		if err = repo.SaveSample(sample); err != nil {
			return err
		}
		return s.audit(repo, actor, "split", "sample", id, requestID, map[string]string{"parts": strconv.Itoa(len(parts)), "loss": loss.String(), "lossReason": reason})
	})
	return created, err
}

type CustodyInput struct {
	SubsampleID    string               `json:"subsampleId"`
	ToUserID       string               `json:"toUserId"`
	FromLocationID string               `json:"fromLocationId"`
	ToLocationID   string               `json:"toLocationId"`
	Action         domain.CustodyAction `json:"action"`
	Condition      string               `json:"condition"`
	Notes          string               `json:"notes"`
}

func (s *Service) CreateCustody(ctx context.Context, actor, requestID string, in CustodyInput) (domain.CustodyTransfer, error) {
	x, err := domain.NewCustodyTransfer(in.SubsampleID, actor, in.ToUserID, in.FromLocationID, in.ToLocationID, in.Action, in.Condition, in.Notes, s.clock())
	if err != nil {
		return x, err
	}
	err = s.store.Transaction(ctx, func(repo ports.Repository) error {
		if _, err := repo.Subsample(in.SubsampleID); err != nil {
			return err
		}
		if err := repo.SaveCustody(x); err != nil {
			return err
		}
		return s.audit(repo, actor, "initiate", "custody", x.ID, requestID, map[string]string{"status": string(x.Status)})
	})
	return x, err
}
func (s *Service) ConfirmCustody(ctx context.Context, actor, id, requestID string) (domain.CustodyTransfer, error) {
	var x domain.CustodyTransfer
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		x, err = repo.Custody(id)
		if err != nil {
			return err
		}
		if err = x.Confirm(actor, s.clock()); err != nil {
			return err
		}
		sub, err := repo.Subsample(x.SubsampleID)
		if err != nil {
			return err
		}
		sub.CurrentHolder = actor
		sub.LocationID = x.ToLocationID
		if err = repo.SaveSubsample(sub); err != nil {
			return err
		}
		if err = repo.SaveCustody(x); err != nil {
			return err
		}
		return s.audit(repo, actor, "confirm", "custody", id, requestID, map[string]string{"status": string(x.Status)})
	})
	return x, err
}
func (s *Service) CreateMethod(ctx context.Context, actor, requestID string, m domain.Method) (domain.Method, error) {
	m.ID = domain.NewID()
	m.CreatedBy = actor
	m.CreatedAt = s.clock()
	m.Status = domain.MethodDraft
	if m.Version == 0 {
		m.Version = 1
	}
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		for _, x := range repo.Methods() {
			if x.Code == m.Code && x.Version == m.Version {
				return domain.ErrConflict
			}
		}
		if err := repo.SaveMethod(m); err != nil {
			return err
		}
		return s.audit(repo, actor, "create", "method", m.ID, requestID, map[string]string{"version": strconv.Itoa(m.Version)})
	})
	return m, err
}
func (s *Service) PublishMethod(ctx context.Context, actor, id, requestID string) (domain.Method, error) {
	var m domain.Method
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		m, err = repo.Method(id)
		if err != nil {
			return err
		}
		if err = m.Publish(s.clock()); err != nil {
			return err
		}
		if err = repo.SaveMethod(m); err != nil {
			return err
		}
		return s.audit(repo, actor, "publish", "method", id, requestID, map[string]string{"status": string(m.Status)})
	})
	return m, err
}

type TaskInput struct {
	RequestID   string `json:"requestId"`
	SubsampleID string `json:"subsampleId"`
	MethodID    string `json:"methodId"`
	AssigneeID  string `json:"assigneeId"`
}

func (s *Service) CreateTask(ctx context.Context, actor, requestID string, in TaskInput) (domain.Task, error) {
	m, err := s.store.Method(in.MethodID)
	if err != nil {
		return domain.Task{}, err
	}
	if m.Status != domain.MethodPublished {
		return domain.Task{}, domain.ErrInvalidTransition
	}
	if _, err = s.store.Subsample(in.SubsampleID); err != nil {
		return domain.Task{}, err
	}
	t := domain.Task{ID: domain.NewID(), RequestID: in.RequestID, SubsampleID: in.SubsampleID, MethodID: m.ID, MethodVersion: m.Version, AssigneeID: in.AssigneeID, Status: domain.TaskPending, CreatedAt: s.clock(), UpdatedAt: s.clock()}
	err = s.store.Transaction(ctx, func(repo ports.Repository) error {
		if err := repo.SaveTask(t); err != nil {
			return err
		}
		return s.audit(repo, actor, "create", "task", t.ID, requestID, map[string]string{"assignee": t.AssigneeID})
	})
	return t, err
}
func (s *Service) StartTask(ctx context.Context, actor, id, requestID string) (domain.Task, error) {
	var task domain.Task
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		task, err = repo.Task(id)
		if err != nil {
			return err
		}
		has := false
		for _, x := range repo.Custodies() {
			if x.SubsampleID == task.SubsampleID && x.ToUserID == actor && x.Status == domain.CustodyConfirmed {
				has = true
			}
		}
		if err = task.Start(actor, has, s.clock()); err != nil {
			return err
		}
		if err = repo.SaveTask(task); err != nil {
			return err
		}
		return s.audit(repo, actor, "start", "task", id, requestID, map[string]string{"status": string(task.Status)})
	})
	return task, err
}
func (s *Service) SaveRound(ctx context.Context, actor, id, requestID string, readings map[string]decimal.Decimal, reason string) (domain.TaskRound, error) {
	var round domain.TaskRound
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		task, err := repo.Task(id)
		if err != nil {
			return err
		}
		if task.Status != domain.TaskInProgress || task.ExecutorID != actor {
			return domain.ErrForbidden
		}
		m, err := repo.Method(task.MethodID)
		if err != nil {
			return err
		}
		for _, f := range m.Fields {
			if f.Type != domain.FieldNumber {
				continue
			}
			v, ok := readings[f.Name]
			if f.Required && !ok {
				return fmt.Errorf("%w: missing %s", domain.ErrValidation, f.Name)
			}
			if ok && f.Min != nil && v.LessThan(*f.Min) {
				return fmt.Errorf("%w: %s below range", domain.ErrValidation, f.Name)
			}
			if ok && f.Max != nil && v.GreaterThan(*f.Max) {
				return fmt.Errorf("%w: %s above range", domain.ErrValidation, f.Name)
			}
		}
		result, err := domain.Calculate(m, readings)
		if err != nil {
			return err
		}
		round = domain.TaskRound{ID: domain.NewID(), TaskID: id, Number: task.CurrentRound, Readings: readings, Result: result, Decision: "pending", Reason: reason, CreatedBy: actor, CreatedAt: s.clock()}
		if err = repo.SaveRound(round); err != nil {
			return err
		}
		return s.audit(repo, actor, "record", "task_round", round.ID, requestID, map[string]string{"result": result.String()})
	})
	return round, err
}
func (s *Service) SubmitReview(ctx context.Context, actor, id, requestID string) (domain.Task, error) {
	var task domain.Task
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		task, err = repo.Task(id)
		if err != nil {
			return err
		}
		rounds := repo.Rounds(id)
		if len(rounds) == 0 {
			return fmt.Errorf("%w: readings required", domain.ErrValidation)
		}
		if err = task.SubmitReview(actor, rounds[len(rounds)-1], s.clock()); err != nil {
			return err
		}
		if err = repo.SaveTask(task); err != nil {
			return err
		}
		return s.audit(repo, actor, "submit_review", "task", id, requestID, map[string]string{"status": string(task.Status)})
	})
	return task, err
}
func (s *Service) Review(ctx context.Context, actor, id, requestID, decision, comment string) (domain.Task, error) {
	var task domain.Task
	err := s.store.Transaction(ctx, func(repo ports.Repository) error {
		var err error
		task, err = repo.Task(id)
		if err != nil {
			return err
		}
		if comment == "" {
			return fmt.Errorf("%w: review comment", domain.ErrValidation)
		}
		if decision == "approve" {
			err = task.Approve(actor, s.clock())
		} else if decision == "return" {
			err = task.Reject(actor, s.clock())
		} else {
			return domain.ErrValidation
		}
		if err != nil {
			return err
		}
		if err = repo.SaveTask(task); err != nil {
			return err
		}
		return s.audit(repo, actor, "review_"+decision, "task", id, requestID, map[string]string{"comment": comment, "status": string(task.Status)})
	})
	return task, err
}
func (s *Service) IssueCertificate(ctx context.Context, actor, requestID, request string) (domain.Certificate, string, error) {
	var cert domain.Certificate
	code, err := domain.RandomToken(24)
	if err != nil {
		return cert, "", err
	}
	err = s.store.Transaction(ctx, func(repo ports.Repository) error {
		r, err := repo.Request(request)
		if err != nil {
			return err
		}
		methods := map[string]int{}
		results := map[string]decimal.Decimal{}
		executors := []string{}
		reviewers := []string{}
		count := 0
		for _, t := range repo.Tasks() {
			if t.RequestID != request {
				continue
			}
			count++
			if t.Status != domain.TaskApproved {
				return fmt.Errorf("%w: all tasks must be approved", domain.ErrValidation)
			}
			methods[t.MethodID] = t.MethodVersion
			executors = append(executors, t.ExecutorID)
			rounds := repo.Rounds(t.ID)
			if len(rounds) > 0 {
				results[t.ID] = rounds[len(rounds)-1].Result
			}
		}
		if count == 0 {
			return fmt.Errorf("%w: tasks required", domain.ErrValidation)
		}
		number := fmt.Sprintf("CERT-%s-%04d", s.clock().Format("20060102"), len(repo.Certificates())+1)
		digest := domain.Digest(fmt.Sprintf("%s|%s|%v|%s", number, r.SampleID, results, s.clock().Format(time.RFC3339Nano)))
		cert = domain.Certificate{ID: domain.NewID(), Number: number, RequestID: request, SampleID: r.SampleID, MethodVersions: methods, Results: results, Decision: "conform", ExecutorIDs: executors, ReviewerIDs: reviewers, Status: domain.CertificateIssued, VerificationCode: code, Digest: digest, IssuedBy: actor, IssuedAt: s.clock()}
		if err = repo.SaveCertificate(cert); err != nil {
			return err
		}
		return s.audit(repo, actor, "issue", "certificate", cert.ID, requestID, map[string]string{"number": number, "digest": digest})
	})
	return cert, code, err
}
func (s *Service) Store() ports.Store { return s.store }
