package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pterm/pterm"
	pb "github.com/chalkan3/sloth-runner/proto"
	"google.golang.org/grpc"
)

// agentRegistryServer implements the AgentRegistry service.
 type agentRegistryServer struct {
	pb.UnimplementedAgentRegistryServer
	mu      sync.Mutex
	agents  map[string]*pb.AgentInfo
	grpcServer *grpc.Server
}

// newAgentRegistryServer creates a new agentRegistryServer.
func newAgentRegistryServer() *agentRegistryServer {
	return &agentRegistryServer{
		agents: make(map[string]*pb.AgentInfo),
	}
}

// RegisterAgent registers a new agent.
func (s *agentRegistryServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pterm.Success.Printf("Agent registered: %s at %s\n", req.AgentName, req.AgentAddress)
	s.agents[req.AgentName] = &pb.AgentInfo{
		AgentName:    req.AgentName,
		AgentAddress: req.AgentAddress,
	}

	return &pb.RegisterAgentResponse{Success: true, Message: "Agent registered successfully"}, nil
}

// ListAgents lists all registered agents.
func (s *agentRegistryServer) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	pterm.Info.Println("Listing registered agents")
	s.mu.Lock()
	defer s.mu.Unlock()

	var agents []*pb.AgentInfo
	for _, agent := range s.agents {
		// Determine agent status based on last heartbeat
		status := "Inactive"
		pterm.Debug.Printf("Agent %s: LastHeartbeat=%d, CurrentTime=%d, Diff=%d\n", agent.AgentName, agent.LastHeartbeat, time.Now().Unix(), time.Now().Unix()-agent.LastHeartbeat)
		if time.Now().Unix()-agent.LastHeartbeat < 60 { // Agent considered active if heartbeat within last 60 seconds
			status = "Active"
		}
		agents = append(agents, &pb.AgentInfo{
			AgentName:     agent.AgentName,
			AgentAddress:  agent.AgentAddress,
			LastHeartbeat: agent.LastHeartbeat,
			Status:        status,
		})
	}

	return &pb.ListAgentsResponse{Agents: agents}, nil
}

// Heartbeat updates the last heartbeat timestamp for an agent.
func (s *agentRegistryServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent, ok := s.agents[req.AgentName]; ok {
		agent.LastHeartbeat = time.Now().Unix()
		pterm.Debug.Printf("Heartbeat received from agent: %s\n", req.AgentName)
		return &pb.HeartbeatResponse{Success: true, Message: "Heartbeat received"}, nil
	}
	return &pb.HeartbeatResponse{Success: false, Message: "Agent not found"}, nil
}

// ExecuteCommand executes a command on a remote agent.
func (s *agentRegistryServer) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	s.mu.Lock()
	agent, ok := s.agents[req.AgentName]
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("agent not found: %s", req.AgentName)
	}

	conn, err := grpc.Dial(agent.AgentAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentClient(conn)
	resp, err := client.RunCommand(ctx, &pb.RunCommandRequest{Command: req.Command})
	if err != nil {
		return nil, fmt.Errorf("failed to run command on agent: %v", err)
	}

	return &pb.ExecuteCommandResponse{
		Success: resp.Success,
		Stdout:  resp.Stdout,
		Stderr:  resp.Stderr,
		Error:   resp.Error,
	}, nil
}

// StopAgent stops a remote agent.
func (s *agentRegistryServer) StopAgent(ctx context.Context, req *pb.StopAgentRequest) (*pb.StopAgentResponse, error) {
	s.mu.Lock()
	agent, ok := s.agents[req.AgentName]
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("agent not found: %s", req.AgentName)
	}

	conn, err := grpc.Dial(agent.AgentAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentClient(conn)
	_, err = client.Shutdown(ctx, &pb.ShutdownRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to stop agent: %v", err)
	}

	return &pb.StopAgentResponse{Success: true, Message: "Agent stopped successfully"}, nil
}

// Start starts the agent registry server.
func (s *agentRegistryServer) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.grpcServer = grpc.NewServer()
	pb.RegisterAgentRegistryServer(s.grpcServer, s)
	pterm.Info.Printf("Agent registry listening at %v\n", lis.Addr())
	return s.grpcServer.Serve(lis)
}

// Stop stops the agent registry server.
func (s *agentRegistryServer) Stop() {
	s.grpcServer.GracefulStop()
}
