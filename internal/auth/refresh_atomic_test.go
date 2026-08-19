package auth

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"material-lab-platform/internal/domain"
)

func TestRefreshRotationRemainsAtomicUnderConcurrentUse(t *testing.T) {
	service := New("refresh-test-secret", time.Minute, time.Hour)
	user, err := service.CreateUser("operator", "Operator", "StrongPassword1!", domain.RoleManager)
	if err != nil {
		t.Fatal(err)
	}
	_, refresh, _, err := service.Login("operator", "StrongPassword1!", "test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	start := make(chan struct{})
	var wait sync.WaitGroup
	for i := 0; i < 2; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			if _, _, refreshErr := service.Refresh(refresh); refreshErr == nil {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful rotations = %d, want 1", successes.Load())
	}
	active := 0
	for _, session := range service.Sessions(user.ID) {
		if session.RevokedAt.IsZero() {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("active sessions = %d, want 1", active)
	}
}

func TestRefreshRejectsMalformedAndInactiveSessionState(t *testing.T) {
	service := New("refresh-test-secret", time.Minute, time.Hour)
	user, err := service.CreateUser("operator", "Operator", "StrongPassword1!", domain.RoleManager)
	if err != nil {
		t.Fatal(err)
	}
	_, refresh, _, err := service.Login("operator", "StrongPassword1!", "test", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = service.Refresh("missing-separator"); err == nil {
		t.Fatal("malformed refresh accepted")
	}
	if _, _, err = service.Refresh(refresh + ".suffix"); err == nil {
		t.Fatal("multi-part refresh accepted")
	}
	service.mu.Lock()
	changed := service.users[user.ID]
	changed.Active = false
	service.users[user.ID] = changed
	service.mu.Unlock()
	if _, _, err = service.Refresh(refresh); err == nil {
		t.Fatal("inactive user refresh accepted")
	}
}
