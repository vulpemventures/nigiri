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
	GetPortsForService(composePath string, serviceName string) ([]string, error)
	IsContainerRunning(containerName string) (bool, error)
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

	// Extract endpoints
	endpoints := make(map[string]string)
	for service, details := range config.Services {
		for _, port := range details.Ports {
			parts := strings.Split(port, ":")
			if len(parts) >= 2 {
				// Use the first port mapping found
				endpoints[service] = fmt.Sprintf("localhost:%s", parts[0])
				break
			}
		}
	}

	return endpoints, nil
}

// GetPortsForService returns all exposed ports for a specific service
func (c *DefaultClient) GetPortsForService(composePath string, serviceName string) ([]string, error) {
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

	// Get service details
	service, exists := config.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found in compose file", serviceName)
	}

	// Extract host ports
	var ports []string
	for _, port := range service.Ports {
		parts := strings.Split(port, ":")
		if len(parts) >= 2 {
			ports = append(ports, parts[0])
		}
	}

	return ports, nil
}

// IsContainerRunning checks if a container is running
func (c *DefaultClient) IsContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return false, nil // Container doesn't exist or other error
	}
	return strings.TrimSpace(string(output)) == "true", nil
}
