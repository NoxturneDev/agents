package daemon

import (
	"context"
	"testing"
	"time"

	"antigravity/pb"
)

func TestLockServer_AcquireAndRelease(t *testing.T) {
	server := NewLockServer()

	// 1. Initial acquire
	resp, err := server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-1",
	})
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if !resp.Granted {
		t.Errorf("expected lock to be granted, got false with message: %s", resp.Message)
	}

	// 2. Try acquiring same lock with different agent
	resp, err = server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-2",
	})
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if resp.Granted {
		t.Errorf("expected lock to be denied, but got granted=true")
	}
	if resp.Message != "LOCKED_BY_agent-1" {
		t.Errorf("expected message LOCKED_BY_agent-1, got: %s", resp.Message)
	}

	// 3. Release lock with wrong agent
	resp, err = server.ReleaseLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-2",
	})
	if err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}
	if resp.Granted {
		t.Errorf("expected release to fail, but got granted=true")
	}

	// 4. Release lock with owner agent
	resp, err = server.ReleaseLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-1",
	})
	if err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}
	if !resp.Granted {
		t.Errorf("expected release to succeed, got false with message: %s", resp.Message)
	}
}

func TestLockServer_TTLExpiration(t *testing.T) {
	server := NewLockServer()

	// 1. Acquire lock
	resp, err := server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-1",
	})
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if !resp.Granted {
		t.Fatalf("expected lock to be granted")
	}

	// 2. Access the internal locks map and force expiration
	server.mu.Lock()
	state, ok := server.locks["test.txt"]
	if !ok {
		server.mu.Unlock()
		t.Fatalf("lock not found in server map")
	}
	state.ExpiresAt = time.Now().Add(-1 * time.Second) // Set to 1 second ago
	server.locks["test.txt"] = state
	server.mu.Unlock()

	// 3. Try acquiring lock with agent-2. It should succeed because lock is expired
	resp, err = server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "test.txt",
		AgentId:    "agent-2",
	})
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if !resp.Granted {
		t.Errorf("expected lock to be granted to agent-2 after expiration, got false")
	}
	if resp.Message != "LOCKED" {
		t.Errorf("expected message LOCKED, got: %s", resp.Message)
	}
}
