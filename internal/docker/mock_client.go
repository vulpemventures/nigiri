package docker

import (
	"os/exec"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	RunComposeFunc        func(composePath string, args ...string) *exec.Cmd
	GetEndpointsFunc     func(composePath string) (map[string]string, error)
	GetPortsForServiceFunc func(composePath string, serviceName string) ([]string, error)
	IsContainerRunningFunc func(containerName string) (bool, error)

	// For tracking test interactions
	commands []struct {
		composePath string
		args        []string
	}
	mockCmd *exec.Cmd
}

func NewMockClient() *MockClient {
	return &MockClient{
		commands: make([]struct {
			composePath string
			args        []string
		}, 0),
	}
}

func (m *MockClient) RunCompose(composePath string, args ...string) *exec.Cmd {
	// Store the command for later inspection
	m.commands = append(m.commands, struct {
		composePath string
		args        []string
	}{
		composePath: composePath,
		args:        args,
	})

	if m.RunComposeFunc != nil {
		return m.RunComposeFunc(composePath, args...)
	}

	if m.mockCmd != nil {
		return m.mockCmd
	}

	// Return a mock command that does nothing
	return exec.Command("echo", "mock command")
}

func (m *MockClient) GetEndpoints(composePath string) (map[string]string, error) {
	if m.GetEndpointsFunc != nil {
		return m.GetEndpointsFunc(composePath)
	}
	// Default mock endpoints
	return map[string]string{
		"bitcoin":    "localhost:18443",
		"electrs":    "localhost:3002",
		"esplora":    "localhost:3000",
		"chopsticks": "localhost:3000",
		"ark":        "localhost:7070",
	}, nil
}

func (m *MockClient) GetPortsForService(composePath string, serviceName string) ([]string, error) {
	if m.GetPortsForServiceFunc != nil {
		return m.GetPortsForServiceFunc(composePath, serviceName)
	}
	// Default mock ports
	switch serviceName {
	case "bitcoin":
		return []string{"18443"}, nil
	case "electrs":
		return []string{"3002"}, nil
	case "esplora":
		return []string{"3000"}, nil
	case "chopsticks":
		return []string{"3000"}, nil
	case "ark":
		return []string{"7070"}, nil
	default:
		return []string{}, nil
	}
}

// Test helper methods
func (m *MockClient) GetLastCommand() (string, []string, bool) {
	if len(m.commands) == 0 {
		return "", nil, false
	}
	last := m.commands[len(m.commands)-1]
	return last.composePath, last.args, true
}

func (m *MockClient) ClearCommands() {
	m.commands = m.commands[:0]
}

func (m *MockClient) SetMockCmd(cmd *exec.Cmd) {
	m.mockCmd = cmd
}

// Alias for backward compatibility
func (m *MockClient) MockCmd() *exec.Cmd {
	return m.mockCmd
}

func (m *MockClient) IsContainerRunning(containerName string) (bool, error) {
	if m.IsContainerRunningFunc != nil {
		return m.IsContainerRunningFunc(containerName)
	}
	// Default mock behavior - return true for testing
	return true, nil
}
