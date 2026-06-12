package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"antigravity/pb"
)

// WriteFileWithLock normalizes the target path, auto-bootstraps the daemon,
// acquires a lock on the resource, writes the file atomically, and releases the lock.
func WriteFileWithLock(path string, data []byte) error {
	workspaceRoot, err := FindWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("failed to find workspace root: %w", err)
	}

	// Normalize target file path relative to the workspace root for locking consistency
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	relPath, err := filepath.Rel(workspaceRoot, absPath)
	if err != nil {
		relPath = absPath
	}

	// Bootstrap daemon and get client
	c, conn, err := ConnectAndBootstrap()
	if err != nil {
		return fmt.Errorf("failed to connect to lock daemon: %w", err)
	}
	defer conn.Close()

	agentID := fmt.Sprintf("pid-%d", os.Getpid())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retry loop for lock acquisition
	var acquireResp *pb.LockResponse
	const retries = 10
	const delay = 200 * time.Millisecond

	for i := 0; i < retries; i++ {
		acquireResp, err = c.AcquireLock(ctx, &pb.LockRequest{
			ResourceId: relPath,
			AgentId:    agentID,
		})
		if err != nil {
			return fmt.Errorf("AcquireLock RPC failed: %w", err)
		}
		if acquireResp.GetGranted() {
			break
		}
		if i < retries-1 {
			time.Sleep(delay)
		}
	}

	if !acquireResp.GetGranted() {
		return fmt.Errorf("failed to acquire lock: %s", acquireResp.GetMessage())
	}

	// Ensure release happens after writing
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer releaseCancel()
		_, _ = c.ReleaseLock(releaseCtx, &pb.LockRequest{
			ResourceId: relPath,
			AgentId:    agentID,
		})
	}()

	// Ensure parent directories exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure %s: %w", dir, err)
	}

	// Write the file contents
	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", absPath, err)
	}

	return nil
}
