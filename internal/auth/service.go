package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"material-lab-platform/internal/domain"
)

type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	TokenDigest string    `json:"-"`
	UserAgent   string    `json:"userAgent"`
	ExpiresAt   time.Time `json:"expiresAt"`
	RevokedAt   time.Time `json:"revokedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
type Claims struct {
	UserID string      `json:"uid"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}
type attempt struct {
	count int
	reset time.Time
}
type Service struct {
	mu                    sync.RWMutex
	secret                []byte
	accessTTL, refreshTTL time.Duration
	users                 map[string]domain.User
	byName                map[string]string
	sessions              map[string]Session
	attempts              map[string]attempt
}

func New(secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL, users: map[string]domain.User{}, byName: map[string]string{}, sessions: map[string]Session{}, attempts: map[string]attempt{}}
}
func (s *Service) CreateUser(username, name, password string, role domain.Role) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !role.Valid() {
		return domain.User{}, domain.ErrValidation
	}
	if _, ok := s.byName[username]; ok {
		return domain.User{}, domain.ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{ID: domain.NewID(), Username: username, DisplayName: name, PasswordHash: string(hash), Role: role, Active: true, CreatedAt: time.Now().UTC()}
	s.users[u.ID] = u
	s.byName[username] = u.ID
	return u, nil
}
func (s *Service) Login(username, password, userAgent, remote string) (string, string, domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	a := s.attempts[remote]
	if now.Before(a.reset) && a.count >= 5 {
		return "", "", domain.User{}, fmt.Errorf("rate limited")
	}
	id, ok := s.byName[username]
	u := s.users[id]
	if !ok || !u.Active || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		if now.After(a.reset) {
			a = attempt{reset: now.Add(5 * time.Minute)}
		}
		a.count++
		s.attempts[remote] = a
		return "", "", domain.User{}, errors.New("invalid credentials")
	}
	delete(s.attempts, remote)
	access, err := s.sign(u, now)
	if err != nil {
		return "", "", domain.User{}, err
	}
	refresh, err := domain.RandomToken(32)
	if err != nil {
		return "", "", domain.User{}, err
	}
	session := Session{ID: domain.NewID(), UserID: u.ID, TokenDigest: domain.Digest(refresh), UserAgent: userAgent, ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now}
	s.sessions[session.ID] = session
	return access, session.ID + "." + refresh, u, nil
}
func (s *Service) sign(u domain.User, now time.Time) (string, error) {
	claims := Claims{UserID: u.ID, Role: u.Role, RegisteredClaims: jwt.RegisteredClaims{Subject: u.ID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
func (s *Service) Parse(token string) (domain.User, error) {
	claims := Claims{}
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("signing method")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return domain.User{}, errors.New("invalid token")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[claims.UserID]
	if !ok || !u.Active {
		return domain.User{}, errors.New("inactive user")
	}
	return u, nil
}
func (s *Service) Refresh(raw string) (string, string, error) {
	sid, token, err := domain.SplitOpaqueToken(raw)
	if err != nil {
		return "", "", errors.New("invalid refresh")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sid]
	now := time.Now().UTC()
	if !ok || !session.RevokedAt.IsZero() || !now.Before(session.ExpiresAt) || !domain.SecureDigestEqual(token, session.TokenDigest) {
		return "", "", errors.New("invalid refresh")
	}
	u, ok := s.users[session.UserID]
	if !ok || !u.Active {
		return "", "", errors.New("invalid refresh")
	}
	session.RevokedAt = now
	s.sessions[sid] = session
	access, err := s.sign(u, now)
	if err != nil {
		return "", "", err
	}
	next, err := domain.RandomToken(32)
	if err != nil {
		return "", "", err
	}
	ns := Session{ID: domain.NewID(), UserID: u.ID, TokenDigest: domain.Digest(next), UserAgent: session.UserAgent, ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now}
	s.sessions[ns.ID] = ns
	return access, ns.ID + "." + next, nil
}
func (s *Service) Logout(sessionID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return domain.ErrNotFound
	}
	if session.UserID != userID {
		return domain.ErrForbidden
	}
	session.RevokedAt = time.Now().UTC()
	s.sessions[sessionID] = session
	return nil
}
func (s *Service) Sessions(userID string) []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Session{}
	for _, x := range s.sessions {
		if x.UserID == userID {
			out = append(out, x)
		}
	}
	return out
}
func (s *Service) User(id string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}
func (s *Service) Users() []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}
