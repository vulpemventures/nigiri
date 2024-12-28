package docker

import (
	"os/exec"
)

// Client defines the interface for Docker operations
type Client interface {
	RunCompose(composePath string, args ...string) *exec.Cmd
}

// DefaultClient is the default implementation using real Docker commands
type DefaultClient struct{}

func NewDefaultClient() *DefaultClient {
	return &DefaultClient{}
}

func (c *DefaultClient) RunCompose(composePath string, args ...string) *exec.Cmd {
	var cmd *exec.Cmd

	// Try to find docker-compose binary
	_, err := exec.LookPath("docker-compose")
	if err == nil {
		// docker-compose found, use it
		cmd = exec.Command("docker-compose", append([]string{"-f", composePath}, args...)...)
	} else {
		// docker-compose not found, fallback to 'docker compose'
		cmd = exec.Command("docker", append([]string{"compose", "-f", composePath}, args...)...)
	}

	return cmd
}
