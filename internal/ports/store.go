package ports

import (
	"context"
	"material-lab-platform/internal/domain"
)

type Store interface {
	Transaction(context.Context, func(Repository) error) error
	Repository
}
type Repository interface {
	SaveRequest(domain.InspectionRequest) error
	Request(string) (domain.InspectionRequest, error)
	Requests() []domain.InspectionRequest
	NextReference() string
	SaveSample(domain.Sample) error
	Sample(string) (domain.Sample, error)
	Samples() []domain.Sample
	SaveSubsample(domain.Subsample) error
	Subsamples(string) []domain.Subsample
	Subsample(string) (domain.Subsample, error)
	SaveCustody(domain.CustodyTransfer) error
	Custody(string) (domain.CustodyTransfer, error)
	Custodies() []domain.CustodyTransfer
	SaveMethod(domain.Method) error
	Method(string) (domain.Method, error)
	Methods() []domain.Method
	SaveTask(domain.Task) error
	Task(string) (domain.Task, error)
	Tasks() []domain.Task
	SaveRound(domain.TaskRound) error
	Rounds(string) []domain.TaskRound
	SaveRetest(domain.RetestRequest) error
	Retests(string) []domain.RetestRequest
	SaveCertificate(domain.Certificate) error
	Certificate(string) (domain.Certificate, error)
	CertificateByCode(string) (domain.Certificate, error)
	Certificates() []domain.Certificate
	AppendAudit(domain.AuditEvent) error
	Audits() []domain.AuditEvent
}
