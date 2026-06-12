package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"antigravity/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CheckDaemonHealth checks if the daemon is running and responds to Ping.
func CheckDaemonHealth(socketPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", addr)
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return false
	}
	defer conn.Close()

	c := pb.NewLockManagerClient(conn)
	resp, err := c.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return false
	}
	return resp.GetStatus() == "ALIVE"
}

// ConnectAndBootstrap ensures the daemon is running (spawning it if necessary)
// and returns a client connection to it.
func ConnectAndBootstrap() (pb.LockManagerClient, *grpc.ClientConn, error) {
	workspaceRoot, err := FindWorkspaceRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find workspace root: %w", err)
	}

	socketPath := filepath.Join(workspaceRoot, ".agents", "antigravity.sock")

	// Pre-flight check
	if CheckDaemonHealth(socketPath) {
		conn, err := dial(socketPath)
		if err != nil {
			return nil, nil, err
		}
		return pb.NewLockManagerClient(conn), conn, nil
	}

	// Failure case: Dead or missing socket.
	// 1. Remove existing socket file if any
	_ = os.Remove(socketPath)

	// Ensure the parent directory (.agents) exists before starting daemon
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory for socket: %w", err)
	}

	// 2. Spawn the daemon in the background
	execPath, err := os.Executable()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current executable: %w", err)
	}

	cmd := exec.Command(execPath, "daemon")
	cmd.Dir = workspaceRoot
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	// Detach standard streams so child process survives terminal exits and does not print to stdout
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start daemon process: %w", err)
	}

	// 3. Wait 100ms
	time.Sleep(100 * time.Millisecond)

	// 4. Retry Ping
	if !CheckDaemonHealth(socketPath) {
		return nil, nil, fmt.Errorf("failed to contact daemon after spawn retry")
	}

	conn, err := dial(socketPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial daemon after successful spawn: %w", err)
	}

	return pb.NewLockManagerClient(conn), conn, nil
}

func dial(socketPath string) (*grpc.ClientConn, error) {
	return grpc.Dial(socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", addr)
		}),
	)
}
