package docker

import (
	"os/exec"
)

// MockClient is a mock implementation of the Docker client for testing
type MockClient struct {
	// Commands stores the commands that were executed
	Commands []struct {
		ComposePath string
		Args        []string
	}
	// MockCmd is the command that will be returned by RunCompose
	MockCmd *exec.Cmd
}

func NewMockClient() *MockClient {
	return &MockClient{
		Commands: make([]struct {
			ComposePath string
			Args        []string
		}, 0),
	}
}

func (m *MockClient) RunCompose(composePath string, args ...string) *exec.Cmd {
	// Store the command for later inspection
	m.Commands = append(m.Commands, struct {
		ComposePath string
		Args        []string
	}{
		ComposePath: composePath,
		Args:        args,
	})

	if m.MockCmd != nil {
		return m.MockCmd
	}

	// Return a mock command that does nothing
	return exec.Command("echo", "mock command")
}

// GetLastCommand returns the last command that was executed
func (m *MockClient) GetLastCommand() (string, []string, bool) {
	if len(m.Commands) == 0 {
		return "", nil, false
	}
	last := m.Commands[len(m.Commands)-1]
	return last.ComposePath, last.Args, true
}

// ClearCommands clears the stored commands
func (m *MockClient) ClearCommands() {
	m.Commands = m.Commands[:0]
}
