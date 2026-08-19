package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shopspring/decimal"
	"material-lab-platform/internal/application"
	"material-lab-platform/internal/auth"
	"material-lab-platform/internal/config"
	"material-lab-platform/internal/domain"
)

type contextKey string

const userKey contextKey = "user"

type Server struct {
	app   *application.Service
	auth  *auth.Service
	cfg   config.Config
	ready atomic.Bool
}
type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId"`
	Fields    map[string]string `json:"fields,omitempty"`
}
type envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
	Meta  any       `json:"meta,omitempty"`
}

func New(app *application.Service, authService *auth.Service, cfg config.Config) http.Handler {
	s := &Server{app: app, auth: authService, cfg: cfg}
	s.ready.Store(true)
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, s.securityHeaders, s.logging, s.timeout, s.limitBody)
	r.Get("/health/live", s.live)
	r.Get("/health/ready", s.readiness)
	r.Get("/metrics", s.metrics)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", s.login)
		r.Post("/auth/refresh", s.refresh)
		r.Get("/public/certificates/verify", s.publicVerify)
		r.Group(func(r chi.Router) {
			r.Use(s.authenticate)
			r.Post("/auth/logout", s.logout)
			r.Get("/me", s.me)
			r.Get("/sessions", s.sessions)
			r.Delete("/sessions/{id}", s.deleteSession)
			r.Get("/users", s.users)
			r.Get("/commissions", s.requests)
			r.Post("/commissions", s.createRequest)
			r.Get("/commissions/{id}", s.request)
			r.Patch("/commissions/{id}", s.updateRequest)
			r.Post("/commissions/{id}/submit", s.submitRequest)
			r.Get("/samples", s.samples)
			r.Get("/samples/{id}", s.sample)
			r.Post("/samples/{id}/splits", s.splitSample)
			r.Get("/custody-transfers", s.custodies)
			r.Post("/custody-transfers", s.createCustody)
			r.Post("/custody-transfers/{id}/confirm", s.confirmCustody)
			r.Get("/methods", s.methods)
			r.Post("/methods", s.createMethod)
			r.Get("/methods/{id}", s.method)
			r.Post("/methods/{id}/validate", s.validateMethod)
			r.Post("/methods/{id}/publish", s.publishMethod)
			r.Get("/tasks", s.tasks)
			r.Post("/tasks", s.createTask)
			r.Get("/tasks/{id}", s.task)
			r.Post("/tasks/{id}/start", s.startTask)
			r.Post("/tasks/{id}/rounds", s.saveRound)
			r.Post("/tasks/{id}/submit-review", s.submitReview)
			r.Post("/tasks/{id}/retests", s.createRetest)
			r.Get("/reviews", s.reviews)
			r.Get("/reviews/{id}", s.review)
			r.Post("/reviews/{id}/decision", s.reviewDecision)
			r.Get("/certificates", s.certificates)
			r.Get("/certificates/{id}", s.certificate)
			r.Post("/certificates/issue", s.issueCertificate)
			r.Get("/audits", s.audits)
		})
	})
	return r
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http_request", "method", r.Method, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()), "duration_ms", time.Since(start).Milliseconds())
	})
}
func (s *Server) timeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 20*time.Second, `{"error":{"code":"timeout","message":"request timed out"}}`)
}
func (s *Server) limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			s.fail(w, r, http.StatusUnauthorized, "unauthorized", "access token required", nil)
			return
		}
		user, err := s.auth.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			s.fail(w, r, http.StatusUnauthorized, "unauthorized", "access token invalid or expired", nil)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}
func current(r *http.Request) domain.User { return r.Context().Value(userKey).(domain.User) }
func (s *Server) require(w http.ResponseWriter, r *http.Request, roles ...domain.Role) (domain.User, bool) {
	u := current(r)
	for _, role := range roles {
		if u.Role == role {
			return u, true
		}
	}
	s.fail(w, r, http.StatusForbidden, "forbidden", "role is not permitted", nil)
	return u, false
}
func (s *Server) decode(w http.ResponseWriter, r *http.Request, out any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return false
	}
	return true
}
func (s *Server) respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}
func (s *Server) fail(w http.ResponseWriter, r *http.Request, status int, code, msg string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: &apiError{Code: code, Message: msg, RequestID: middleware.GetReqID(r.Context()), Fields: fields}})
}
func (s *Server) handle(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		s.fail(w, r, 404, "not_found", err.Error(), nil)
	case errors.Is(err, domain.ErrForbidden):
		s.fail(w, r, 403, "forbidden", err.Error(), nil)
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrInvalidTransition):
		s.fail(w, r, 409, "conflict", err.Error(), nil)
	case errors.Is(err, domain.ErrValidation):
		s.fail(w, r, 422, "validation", err.Error(), nil)
	default:
		s.fail(w, r, 500, "internal", "internal server error", nil)
	}
}
func requestID(r *http.Request) string { return middleware.GetReqID(r.Context()) }

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, map[string]string{"status": "alive"})
}
func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	if !s.ready.Load() {
		s.fail(w, r, 503, "not_ready", "service not ready", nil)
		return
	}
	s.respond(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("# HELP material_lab_up Whether the service is up.\n# TYPE material_lab_up gauge\nmaterial_lab_up 1\n"))
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	access, refresh, user, err := s.auth.Login(in.Username, in.Password, r.UserAgent(), host)
	if err != nil {
		s.fail(w, r, 401, "invalid_credentials", err.Error(), nil)
		return
	}
	s.respond(w, 200, map[string]any{"accessToken": access, "refreshToken": refresh, "expiresIn": int(s.cfg.AccessTTL.Seconds()), "user": user})
}
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refreshToken"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	access, refresh, err := s.auth.Refresh(in.RefreshToken)
	if err != nil {
		s.fail(w, r, 401, "invalid_refresh", err.Error(), nil)
		return
	}
	s.respond(w, 200, map[string]string{"accessToken": access, "refreshToken": refresh})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SessionID string `json:"sessionId"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	if err := s.auth.Logout(in.SessionID, current(r).ID); err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, map[string]bool{"revoked": true})
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, map[string]any{"user": current(r), "laboratory": s.cfg.LabName, "timezone": s.cfg.Timezone})
}
func (s *Server) sessions(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, s.auth.Sessions(current(r).ID))
}
func (s *Server) deleteSession(w http.ResponseWriter, r *http.Request) {
	if err := s.auth.Logout(chi.URLParam(r, "id"), current(r).ID); err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, map[string]bool{"revoked": true})
}
func (s *Server) users(w http.ResponseWriter, r *http.Request) { s.respond(w, 200, s.auth.Users()) }

func (s *Server) requests(w http.ResponseWriter, r *http.Request) {
	items := s.app.Store().Requests()
	status := r.URL.Query().Get("status")
	filtered := items[:0]
	for _, x := range items {
		if status == "" || string(x.Status) == status {
			filtered = append(filtered, x)
		}
	}
	limit := parseLimit(r, 50)
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	s.respond(w, 200, map[string]any{"items": filtered, "nextCursor": ""})
}
func (s *Server) request(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.Store().Request(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) createRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleRegistrar, domain.RoleManager)
	if !ok {
		return
	}
	var in application.RequestInput
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.CreateRequest(r.Context(), u.ID, requestID(r), in)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) updateRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleRegistrar, domain.RoleManager)
	if !ok {
		return
	}
	var in application.RequestInput
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.UpdateRequest(r.Context(), u.ID, chi.URLParam(r, "id"), requestID(r), in)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) submitRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleRegistrar, domain.RoleManager)
	if !ok {
		return
	}
	var in struct {
		TotalQuantity decimal.Decimal `json:"totalQuantity"`
		Unit          string          `json:"unit"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	request, sample, err := s.app.SubmitRequest(r.Context(), u.ID, chi.URLParam(r, "id"), requestID(r), in.TotalQuantity, in.Unit)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, map[string]any{"commission": request, "sample": sample})
}
func (s *Server) samples(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, map[string]any{"items": s.app.Store().Samples(), "nextCursor": ""})
}
func (s *Server) sample(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.Store().Sample(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, map[string]any{"sample": x, "subsamples": s.app.Store().Subsamples(x.ID)})
}
func (s *Server) splitSample(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleRegistrar, domain.RoleManager)
	if !ok {
		return
	}
	var in struct {
		Parts      []domain.SplitInput `json:"parts"`
		Loss       decimal.Decimal     `json:"loss"`
		LossReason string              `json:"lossReason"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.SplitSample(r.Context(), u.ID, chi.URLParam(r, "id"), requestID(r), in.Parts, in.Loss, in.LossReason)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}

func (s *Server) custodies(w http.ResponseWriter, r *http.Request) {
	items := s.app.Store().Custodies()
	s.respond(w, 200, map[string]any{"items": items, "nextCursor": ""})
}
func (s *Server) createCustody(w http.ResponseWriter, r *http.Request) {
	var in application.CustodyInput
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.CreateCustody(r.Context(), current(r).ID, requestID(r), in)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) confirmCustody(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.ConfirmCustody(r.Context(), current(r).ID, chi.URLParam(r, "id"), requestID(r))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}

func (s *Server) methods(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, map[string]any{"items": s.app.Store().Methods(), "nextCursor": ""})
}
func (s *Server) method(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.Store().Method(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) createMethod(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleManager)
	if !ok {
		return
	}
	var in domain.Method
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.CreateMethod(r.Context(), u.ID, requestID(r), in)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) validateMethod(w http.ResponseWriter, r *http.Request) {
	m, err := s.app.Store().Method(chi.URLParam(r, "id"))
	if err == nil {
		err = m.Validate()
	}
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, map[string]any{"valid": true, "variables": m.Fields})
}
func (s *Server) publishMethod(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleManager)
	if !ok {
		return
	}
	x, err := s.app.PublishMethod(r.Context(), u.ID, chi.URLParam(r, "id"), requestID(r))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}

func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	items := s.app.Store().Tasks()
	status := r.URL.Query().Get("status")
	filtered := items[:0]
	for _, x := range items {
		if status == "" || string(x.Status) == status {
			filtered = append(filtered, x)
		}
	}
	s.respond(w, 200, map[string]any{"items": filtered, "nextCursor": ""})
}
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleRegistrar, domain.RoleManager)
	if !ok {
		return
	}
	var in application.TaskInput
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.CreateTask(r.Context(), u.ID, requestID(r), in)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) task(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.Store().Task(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	m, _ := s.app.Store().Method(x.MethodID)
	s.respond(w, 200, map[string]any{"task": x, "method": m, "rounds": s.app.Store().Rounds(x.ID), "retests": s.app.Store().Retests(x.ID)})
}
func (s *Server) startTask(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.StartTask(r.Context(), current(r).ID, chi.URLParam(r, "id"), requestID(r))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) saveRound(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Readings map[string]decimal.Decimal `json:"readings"`
		Reason   string                     `json:"reason"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.SaveRound(r.Context(), current(r).ID, chi.URLParam(r, "id"), requestID(r), in.Readings, in.Reason)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) submitReview(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.SubmitReview(r.Context(), current(r).ID, chi.URLParam(r, "id"), requestID(r))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) createRetest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	task, err := s.app.Store().Task(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	if task.ExecutorID != current(r).ID || in.Reason == "" {
		s.handle(w, r, domain.ErrForbidden)
		return
	}
	x := domain.RetestRequest{ID: domain.NewID(), TaskID: task.ID, Round: task.CurrentRound, Reason: in.Reason, Status: "pending", CreatedAt: time.Now().UTC()}
	if err = s.app.Store().SaveRetest(x); err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, x)
}
func (s *Server) reviews(w http.ResponseWriter, r *http.Request) {
	items := []domain.Task{}
	for _, x := range s.app.Store().Tasks() {
		if x.Status == domain.TaskReview {
			items = append(items, x)
		}
	}
	s.respond(w, 200, map[string]any{"items": items, "nextCursor": ""})
}
func (s *Server) review(w http.ResponseWriter, r *http.Request) { s.task(w, r) }
func (s *Server) reviewDecision(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleReviewer, domain.RoleManager)
	if !ok {
		return
	}
	var in struct {
		Decision string `json:"decision"`
		Comment  string `json:"comment"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	x, err := s.app.Review(r.Context(), u.ID, chi.URLParam(r, "id"), requestID(r), in.Decision, in.Comment)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}

func (s *Server) certificates(w http.ResponseWriter, r *http.Request) {
	s.respond(w, 200, map[string]any{"items": s.app.Store().Certificates(), "nextCursor": ""})
}
func (s *Server) certificate(w http.ResponseWriter, r *http.Request) {
	x, err := s.app.Store().Certificate(chi.URLParam(r, "id"))
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 200, x)
}
func (s *Server) issueCertificate(w http.ResponseWriter, r *http.Request) {
	u, ok := s.require(w, r, domain.RoleManager)
	if !ok {
		return
	}
	var in struct {
		RequestID string `json:"requestId"`
	}
	if !s.decode(w, r, &in) {
		return
	}
	x, code, err := s.app.IssueCertificate(r.Context(), u.ID, requestID(r), in.RequestID)
	if err != nil {
		s.handle(w, r, err)
		return
	}
	s.respond(w, 201, map[string]any{"certificate": x, "verificationCode": code})
}
func (s *Server) publicVerify(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if len(code) < 32 {
		s.fail(w, r, 404, "not_found", "certificate not found", nil)
		return
	}
	x, err := s.app.Store().CertificateByCode(code)
	if err != nil {
		s.fail(w, r, 404, "not_found", "certificate not found", nil)
		return
	}
	s.respond(w, 200, map[string]any{"valid": x.Status == domain.CertificateIssued, "number": x.Number, "status": x.Status, "issuedAt": x.IssuedAt, "digest": x.Digest})
}
func (s *Server) audits(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.require(w, r, domain.RoleManager); !ok {
		return
	}
	s.respond(w, 200, map[string]any{"items": s.app.Store().Audits(), "nextCursor": ""})
}
func parseLimit(r *http.Request, fallback int) int {
	n, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || n < 1 {
		return fallback
	}
	if n > 100 {
		return 100
	}
	return n
}

var _ = fmt.Sprintf
