package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"antigravity/pb"
)

// RunGitCommandWithLock executes a git command in the specified directory under the
// protection of the exclusive GLOBAL_WORKSPACE lock, ensuring atomic commits.
func RunGitCommandWithLock(dir string, gitArgs ...string) (string, error) {
	// 1. Connect and bootstrap the daemon
	c, conn, err := ConnectAndBootstrap()
	if err != nil {
		return "", fmt.Errorf("git lock bootstrap failed: %w", err)
	}
	defer conn.Close()

	agentID := fmt.Sprintf("pid-%d", os.Getpid())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Acquire lock on GLOBAL_WORKSPACE with retries
	const retries = 15
	const delay = 200 * time.Millisecond
	var acquireResp *pb.LockResponse

	for i := 0; i < retries; i++ {
		acquireResp, err = c.AcquireLock(ctx, &pb.LockRequest{
			ResourceId: "GLOBAL_WORKSPACE",
			AgentId:    "global-" + agentID,
		})
		if err != nil {
			return "", fmt.Errorf("AcquireLock for git failed: %w", err)
		}
		if acquireResp.GetGranted() {
			break
		}
		if i < retries-1 {
			time.Sleep(delay)
		}
	}

	if !acquireResp.GetGranted() {
		return "", fmt.Errorf("failed to acquire GLOBAL_WORKSPACE lock for git: %s", acquireResp.GetMessage())
	}

	// 3. Ensure the lock is released
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer releaseCancel()
		_, _ = c.ReleaseLock(releaseCtx, &pb.LockRequest{
			ResourceId: "GLOBAL_WORKSPACE",
			AgentId:    "global-" + agentID,
		})
	}()

	// 4. Execute the git command
	cmd := exec.Command("git", gitArgs...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git command failed: %w (output: %s)", err, string(output))
	}

	return string(output), nil
}
