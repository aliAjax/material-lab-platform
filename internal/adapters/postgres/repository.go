package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"material-lab-platform/internal/adapters/memory"
	"material-lab-platform/internal/domain"
	"material-lab-platform/internal/ports"
)

// Repository persists each domain aggregate as JSON alongside the normalized
// schema. It keeps the application port independent from pgx row types.
type Repository struct {
	pool *Pool
	mem  *memory.Store
	mu   sync.Mutex
}

func NewRepository(ctx context.Context, pool *Pool) (*Repository, error) {
	repository := &Repository{pool: pool, mem: memory.New()}
	if _, err := pool.Inner().Exec(ctx, `CREATE TABLE IF NOT EXISTS runtime_entities (kind text NOT NULL, id text NOT NULL, payload jsonb NOT NULL, updated_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(kind,id))`); err != nil {
		return nil, fmt.Errorf("ensure runtime repository: %w", err)
	}
	if err := repository.restore(ctx); err != nil {
		return nil, err
	}
	return repository, nil
}

func (r *Repository) restore(ctx context.Context) error {
	loads := []struct {
		kind  string
		apply func([]byte) error
	}{
		{"request", func(data []byte) error {
			var x domain.InspectionRequest
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveRequest(x)
		}},
		{"sample", func(data []byte) error {
			var x domain.Sample
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveSample(x)
		}},
		{"subsample", func(data []byte) error {
			var x domain.Subsample
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveSubsample(x)
		}},
		{"custody", func(data []byte) error {
			var x domain.CustodyTransfer
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveCustody(x)
		}},
		{"method", func(data []byte) error {
			var x domain.Method
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveMethod(x)
		}},
		{"task", func(data []byte) error {
			var x domain.Task
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveTask(x)
		}},
		{"round", func(data []byte) error {
			var x domain.TaskRound
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveRound(x)
		}},
		{"retest", func(data []byte) error {
			var x domain.RetestRequest
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveRetest(x)
		}},
		{"certificate", func(data []byte) error {
			var x domain.Certificate
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.SaveCertificate(x)
		}},
		{"audit", func(data []byte) error {
			var x domain.AuditEvent
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			return r.mem.AppendAudit(x)
		}},
	}
	for _, item := range loads {
		if err := r.load(ctx, item.kind, item.apply); err != nil {
			return err
		}
	}
	return nil
}
func (r *Repository) load(ctx context.Context, kind string, apply func([]byte) error) error {
	rows, err := r.pool.Inner().Query(ctx, `SELECT payload FROM runtime_entities WHERE kind=$1`, kind)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var value []byte
		if err = rows.Scan(&value); err != nil {
			return err
		}
		if err = apply(value); err != nil {
			return err
		}
	}
	return rows.Err()
}
func (r *Repository) persist(kind, id string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.pool.Inner().Exec(context.Background(), `INSERT INTO runtime_entities(kind,id,payload) VALUES($1,$2,$3) ON CONFLICT(kind,id) DO UPDATE SET payload=EXCLUDED.payload,updated_at=now()`, kind, id, data)
	return err
}

func (r *Repository) Transaction(ctx context.Context, fn func(ports.Repository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return fn(r)
}
func (r *Repository) SaveRequest(x domain.InspectionRequest) error {
	if err := r.mem.SaveRequest(x); err != nil {
		return err
	}
	return r.persist("request", x.ID, x)
}
func (r *Repository) Request(id string) (domain.InspectionRequest, error) { return r.mem.Request(id) }
func (r *Repository) Requests() []domain.InspectionRequest                { return r.mem.Requests() }
func (r *Repository) NextReference() string                               { return r.mem.NextReference() }
func (r *Repository) SaveSample(x domain.Sample) error {
	if err := r.mem.SaveSample(x); err != nil {
		return err
	}
	return r.persist("sample", x.ID, x)
}
func (r *Repository) Sample(id string) (domain.Sample, error) { return r.mem.Sample(id) }
func (r *Repository) Samples() []domain.Sample                { return r.mem.Samples() }
func (r *Repository) SaveSubsample(x domain.Subsample) error {
	if err := r.mem.SaveSubsample(x); err != nil {
		return err
	}
	return r.persist("subsample", x.ID, x)
}
func (r *Repository) Subsamples(id string) []domain.Subsample       { return r.mem.Subsamples(id) }
func (r *Repository) Subsample(id string) (domain.Subsample, error) { return r.mem.Subsample(id) }
func (r *Repository) SaveCustody(x domain.CustodyTransfer) error {
	if err := r.mem.SaveCustody(x); err != nil {
		return err
	}
	return r.persist("custody", x.ID, x)
}
func (r *Repository) Custody(id string) (domain.CustodyTransfer, error) { return r.mem.Custody(id) }
func (r *Repository) Custodies() []domain.CustodyTransfer               { return r.mem.Custodies() }
func (r *Repository) SaveMethod(x domain.Method) error {
	if err := r.mem.SaveMethod(x); err != nil {
		return err
	}
	return r.persist("method", x.ID, x)
}
func (r *Repository) Method(id string) (domain.Method, error) { return r.mem.Method(id) }
func (r *Repository) Methods() []domain.Method                { return r.mem.Methods() }
func (r *Repository) SaveTask(x domain.Task) error {
	if err := r.mem.SaveTask(x); err != nil {
		return err
	}
	return r.persist("task", x.ID, x)
}
func (r *Repository) Task(id string) (domain.Task, error) { return r.mem.Task(id) }
func (r *Repository) Tasks() []domain.Task                { return r.mem.Tasks() }
func (r *Repository) SaveRound(x domain.TaskRound) error {
	if err := r.mem.SaveRound(x); err != nil {
		return err
	}
	return r.persist("round", x.ID, x)
}
func (r *Repository) Rounds(id string) []domain.TaskRound { return r.mem.Rounds(id) }
func (r *Repository) SaveRetest(x domain.RetestRequest) error {
	if err := r.mem.SaveRetest(x); err != nil {
		return err
	}
	return r.persist("retest", x.ID, x)
}
func (r *Repository) Retests(id string) []domain.RetestRequest { return r.mem.Retests(id) }
func (r *Repository) SaveCertificate(x domain.Certificate) error {
	if err := r.mem.SaveCertificate(x); err != nil {
		return err
	}
	return r.persist("certificate", x.ID, x)
}
func (r *Repository) Certificate(id string) (domain.Certificate, error) { return r.mem.Certificate(id) }
func (r *Repository) CertificateByCode(code string) (domain.Certificate, error) {
	return r.mem.CertificateByCode(code)
}
func (r *Repository) Certificates() []domain.Certificate { return r.mem.Certificates() }
func (r *Repository) AppendAudit(x domain.AuditEvent) error {
	if err := r.mem.AppendAudit(x); err != nil {
		return err
	}
	return r.persist("audit", x.ID, x)
}
func (r *Repository) Audits() []domain.AuditEvent { return r.mem.Audits() }

var _ ports.Store = (*Repository)(nil)
