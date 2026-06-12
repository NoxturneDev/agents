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

func TestLockServer_ParallelIndependentLocking(t *testing.T) {
	server := NewLockServer()

	// 1. Agent 1 acquires Lock X
	respX, err := server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "resource_X.md",
		AgentId:    "agent-1",
	})
	if err != nil || !respX.Granted {
		t.Fatalf("failed to acquire Lock X: %v, resp: %v", err, respX)
	}

	// 2. Agent 2 acquires Lock Y concurrently (should succeed since resource IDs differ)
	respY, err := server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "resource_Y.md",
		AgentId:    "agent-2",
	})
	if err != nil || !respY.Granted {
		t.Fatalf("failed to acquire Lock Y: %v, resp: %v", err, respY)
	}

	// 3. Double-check that Lock X is still held by Agent 1 (cannot be acquired by Agent 2)
	respX2, err := server.AcquireLock(context.Background(), &pb.LockRequest{
		ResourceId: "resource_X.md",
		AgentId:    "agent-2",
	})
	if err != nil || respX2.Granted {
		t.Fatalf("expected Lock X to be blocked, got: %v", respX2)
	}

	// 4. Release both
	releaseX, err := server.ReleaseLock(context.Background(), &pb.LockRequest{
		ResourceId: "resource_X.md",
		AgentId:    "agent-1",
	})
	if err != nil || !releaseX.Granted {
		t.Fatalf("failed to release Lock X: %v", err)
	}

	releaseY, err := server.ReleaseLock(context.Background(), &pb.LockRequest{
		ResourceId: "resource_Y.md",
		AgentId:    "agent-2",
	})
	if err != nil || !releaseY.Granted {
		t.Fatalf("failed to release Lock Y: %v", err)
	}
}
