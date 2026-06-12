package daemon

import (
	"context"
	"sync"
	"time"

	"antigravity/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LockState struct {
	AgentID   string
	ExpiresAt time.Time
}

type LockServer struct {
	pb.UnimplementedLockManagerServer
	mu    sync.RWMutex
	locks map[string]LockState
}

func NewLockServer() *LockServer {
	return &LockServer{
		locks: make(map[string]LockState),
	}
}

func (s *LockServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Status: "ALIVE",
	}, nil
}

func (s *LockServer) AcquireLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resourceID := req.GetResourceId()
	agentID := req.GetAgentId()

	if resourceID == "" || agentID == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_id and agent_id are required")
	}

	now := time.Now()
	state, exists := s.locks[resourceID]
	if exists {
		if now.After(state.ExpiresAt) {
			// Lock is stale. Evict and grant to new requester.
			s.locks[resourceID] = LockState{
				AgentID:   agentID,
				ExpiresAt: now.Add(5 * time.Minute),
			}
			return &pb.LockResponse{
				Granted: true,
				Message: "LOCKED",
			}, nil
		}
		// Lock exists and is not expired.
		return &pb.LockResponse{
			Granted: false,
			Message: "LOCKED_BY_" + state.AgentID,
		}, nil
	}

	// Lock does not exist.
	s.locks[resourceID] = LockState{
		AgentID:   agentID,
		ExpiresAt: now.Add(5 * time.Minute),
	}
	return &pb.LockResponse{
		Granted: true,
		Message: "LOCKED",
	}, nil
}

func (s *LockServer) ReleaseLock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resourceID := req.GetResourceId()
	agentID := req.GetAgentId()

	if resourceID == "" || agentID == "" {
		return nil, status.Error(codes.InvalidArgument, "resource_id and agent_id are required")
	}

	now := time.Now()
	state, exists := s.locks[resourceID]
	if !exists {
		return &pb.LockResponse{
			Granted: true,
			Message: "NOT_LOCKED",
		}, nil
	}

	if state.AgentID == agentID || now.After(state.ExpiresAt) {
		delete(s.locks, resourceID)
		return &pb.LockResponse{
			Granted: true,
			Message: "RELEASED",
		}, nil
	}

	return &pb.LockResponse{
		Granted: false,
		Message: "LOCKED_BY_" + state.AgentID,
	}, nil
}
