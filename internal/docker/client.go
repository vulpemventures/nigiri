package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"gopkg.in/yaml.v2"
)

// Client defines the interface for Docker operations
type Client interface {
	RunCompose(composePath string, args ...string) *exec.Cmd
	GetEndpoints(composePath string) (map[string]string, error)
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

type ComposeConfig struct {
	Services map[string]struct {
		Ports []string `yaml:"ports"`
	} `yaml:"services"`
}

// GetEndpoints reads the docker-compose file and returns a map of service names to their endpoints
func (c *DefaultClient) GetEndpoints(composePath string) (map[string]string, error) {
	// Read the compose file
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	// Parse YAML
	var config ComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	endpoints := make(map[string]string)

	// Iterate through services and find their ports
	for name, service := range config.Services {
		if len(service.Ports) > 0 {
			// Get the first port mapping
			portMapping := service.Ports[0]
			// Split on : to get published port
			parts := strings.Split(portMapping, ":")
			if len(parts) > 0 {
				port := parts[0]
				endpoints[name] = fmt.Sprintf("localhost:%s", port)
			}
		}
	}

	return endpoints, nil
}
