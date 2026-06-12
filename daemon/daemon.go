package daemon

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"antigravity/pb"

	"google.golang.org/grpc"
)

// Run starts the gRPC LockManager server on a Unix Domain Socket at socketPath.
func Run(socketPath string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(socketPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for socket: %w", err)
	}

	// Remove existing socket file if it exists to handle hard crash residue
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Listen on Unix Domain Socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on UDS %s: %w", socketPath, err)
	}
	defer listener.Close()

	// Write PID file to track the running daemon
	pidPath := socketPath + ".pid"
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}
	defer os.Remove(pidPath)

	// Create and register gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterLockManagerServer(grpcServer, NewLockServer())

	// Handle graceful shutdown signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigs:
		grpcServer.GracefulStop()
		_ = os.Remove(socketPath)
		return fmt.Errorf("daemon received signal %v and stopped gracefully", sig)
	case err := <-errChan:
		_ = os.Remove(socketPath)
		return fmt.Errorf("gRPC server error: %w", err)
	}
}
