package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"antigravity/client"
	"antigravity/daemon"
	"antigravity/pb"
)

func init() {
	// Hook to run the test binary as the UDS daemon when spawned via os.Executable()
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		workspaceRoot, err := client.FindWorkspaceRoot()
		if err != nil {
			os.Exit(1)
		}
		socketPath := filepath.Join(workspaceRoot, ".agents", "antigravity.sock")
		if err := daemon.Run(socketPath); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func cleanupDaemon(t *testing.T, root string) {
	pidPath := filepath.Join(root, ".agents", "antigravity.sock.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Signal(os.Interrupt)
		time.Sleep(50 * time.Millisecond)
		_ = proc.Kill()
	}
	socketPath := filepath.Join(root, ".agents", "antigravity.sock")
	_ = os.Remove(socketPath)
	_ = os.Remove(pidPath)
}

func TestIntegration_WorkspaceRoot(t *testing.T) {
	root, err := client.FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("failed to find workspace root: %v", err)
	}

	// Verify the root contains AGENTS.md
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("expected AGENTS.md to exist in workspace root, got error: %v", err)
	}
}

func TestIntegration_ClientBootstrappingAndLocking(t *testing.T) {
	root, err := client.FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("failed to find workspace root: %v", err)
	}

	// Make sure we clean up the daemon process at the end of the test
	defer cleanupDaemon(t, root)

	socketPath := filepath.Join(root, ".agents", "antigravity.sock")

	// Ensure socket is removed before starting to force bootstrapping
	_ = os.Remove(socketPath)

	// Call ConnectAndBootstrap, spawning the daemon
	c, conn, err := client.ConnectAndBootstrap()
	if err != nil {
		t.Fatalf("ConnectAndBootstrap failed: %v", err)
	}
	defer conn.Close()

	// Verify the daemon is alive
	if !client.CheckDaemonHealth(socketPath) {
		t.Errorf("expected daemon to be healthy and alive")
	}

	// Test acquiring a lock
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	acquireResp, err := c.AcquireLock(ctx, &pb.LockRequest{
		ResourceId: "test-resource.md",
		AgentId:    "test-agent",
	})
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if !acquireResp.Granted {
		t.Errorf("expected lock to be granted, got false")
	}

	// Try acquiring again with a different agent (should fail)
	acquireResp2, err := c.AcquireLock(ctx, &pb.LockRequest{
		ResourceId: "test-resource.md",
		AgentId:    "other-agent",
	})
	if err != nil {
		t.Fatalf("AcquireLock 2 failed: %v", err)
	}
	if acquireResp2.Granted {
		t.Errorf("expected lock to be denied for other-agent")
	}

	// Release lock
	releaseResp, err := c.ReleaseLock(ctx, &pb.LockRequest{
		ResourceId: "test-resource.md",
		AgentId:    "test-agent",
	})
	if err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}
	if !releaseResp.Granted {
		t.Errorf("expected release to succeed")
	}
}

func TestIntegration_WriteFileWithLock(t *testing.T) {
	root, err := client.FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("failed to find workspace root: %v", err)
	}

	defer cleanupDaemon(t, root)

	testFile := filepath.Join(root, ".agents", "test_file_lock.txt")
	defer os.Remove(testFile)

	content := []byte("hello lock manager")

	err = client.WriteFileWithLock(testFile, content)
	if err != nil {
		t.Fatalf("WriteFileWithLock failed: %v", err)
	}

	// Verify content was written
	written, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	if string(written) != string(content) {
		t.Errorf("expected '%s', got '%s'", content, written)
	}
}
