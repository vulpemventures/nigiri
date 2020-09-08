package controller

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/vulpemventures/nigiri/cli/config"
	"github.com/vulpemventures/nigiri/cli/constants"
)

var Services = map[string]bool{
	"node":       true,
	"esplora":    true,
	"electrs":    true,
	"chopsticks": true,
}

// Controller implements useful functions to securely parse flags provided at run-time
// and to interact with the resources used by Nigiri:
//   	* docker
//		* .env for docker-compose
//		* nigiri.config.json config file
type Controller struct {
	config *config.Config
	parser *Parser
	docker *Docker
	env    *Env
}

// NewController returns a new Controller instance or error
func NewController() (*Controller, error) {
	c := &Controller{}

	dockerClient := &Docker{}
	if err := dockerClient.New(); err != nil {
		return nil, err
	}
	c.env = &Env{}
	c.parser = newParser(Services)
	c.docker = dockerClient
	c.config = &config.Config{}
	return c, nil
}

// ParseNetwork checks if a valid network has been provided
func (c *Controller) ParseNetwork(network string) error {
	return c.parser.parseNetwork(network)
}

// ParseDatadir checks if a valid datadir has been provided
func (c *Controller) ParseDatadir(path string) error {
	return c.parser.parseDatadir(path)
}

// ParseEnv checks if a valid JSON format for docker compose has been provided
func (c *Controller) ParseEnv(env string) (string, error) {
	return c.parser.parseEnvJSON(env)
}

// ParseServiceName checks if a valid service has been provided
func (c *Controller) ParseServiceName(name string) error {
	return c.parser.parseServiceName(name)
}

// IsNigiriRunning checks if nigiri is running by looking if the bitcoin
// services are in the list of docker running containers
func (c *Controller) IsNigiriRunning() (bool, error) {
	if !c.docker.isDockerRunning() {
		return false, constants.ErrDockerNotRunning
	}
	return c.docker.isNigiriRunning(), nil
}

// IsNigiriStopped checks that nigiri is not actually running and that
// the bitcoin services appear in the list of non running containers
func (c *Controller) IsNigiriStopped() (bool, error) {
	if !c.docker.isDockerRunning() {
		return false, constants.ErrDockerNotRunning
	}
	return c.docker.isNigiriStopped(), nil
}

// WriteComposeEnvironment creates a .env in datadir used by
// the docker-compose YAML file resource
func (c *Controller) WriteComposeEnvironment(datadir, env string) error {
	return c.env.writeEnvForCompose(datadir, env)
}

// ReadComposeEnvironment reads from .env and returns it as a useful type
func (c *Controller) ReadComposeEnvironment(datadir string) (map[string]interface{}, error) {
	return c.env.readEnvForCompose(datadir)
}

// LoadComposeEnvironment returns an os.Environ created from datadir/.env resource
func (c *Controller) LoadComposeEnvironment(datadir string) []string {
	return c.env.load(datadir)
}

// WriteConfigFile writes the configuration handled by the underlying viper
// into the file at filedir path
func (c *Controller) WriteConfigFile(filedir string) error {
	return c.config.WriteConfig(filedir)
}

// ReadConfigFile reads the configuration of the file at filedir path
func (c *Controller) ReadConfigFile(filedir string) error {
	return c.config.ReadFromFile(filedir)
}

// GetConfigBoolField returns a bool field of the config file
func (c *Controller) GetConfigBoolField(field string) bool {
	return c.config.GetBool(field)
}

// GetConfigStringField returns a string field of the config file
func (c *Controller) GetConfigStringField(field string) string {
	return c.config.GetString(field)
}

// NewDatadirFromDefault copies the default ~/.nigiri at the desidered path
// and cleans the docker volumes to make a fresh Nigiri instance
func (c *Controller) NewDatadirFromDefault(datadir string) error {
	defaultDatadir := c.config.GetPath()
	cmd := exec.Command("cp", "-R", filepath.Join(defaultDatadir, "resources"), datadir)
	if err := cmd.Run(); err != nil {
		return err
	}
	c.CleanResourceVolumes(datadir)
	return nil
}

// GetResourcePath returns the absolute path of the requested resource
func (c *Controller) GetResourcePath(datadir, resource string) string {
	if resource == "compose" {
		network := c.config.GetString(constants.Network)
		if c.config.GetBool(constants.AttachLiquid) {
			network += "-liquid"
		}
		return filepath.Join(datadir, "resources", fmt.Sprintf("docker-compose-%s.yml", network))
	}
	if resource == "env" {
		return filepath.Join(datadir, ".env")
	}
	if resource == "config" {
		return filepath.Join(datadir, "nigiri.config.json")
	}
	return ""
}

// CleanResourceVolumes recursively deletes the content of the
// docker volumes in the resource path
func (c *Controller) CleanResourceVolumes(datadir string) error {
	network := c.config.GetString(constants.Network)
	attachLiquid := c.config.GetBool(constants.AttachLiquid)
	if attachLiquid {
		network = fmt.Sprintf("liquid%s", network)
	}
	volumedir := filepath.Join(datadir, "resources", "volumes", network)

	return c.docker.cleanVolumes(volumedir)
}

// GetDefaultDatadir returns the absolute path of Nigiri default directory
func (c *Controller) GetDefaultDatadir() string {
	return c.config.GetPath()
}

// GetServiceName returns the right name of the requested service
// If requesting a name for a Liquid service, then the suffix -liquid
// is appended to the canonical name except for "node" that is mapped
// to either "bitcoin" or "liquid"
func (c *Controller) GetServiceName(name string, liquid bool) string {
	service := name
	if service == "node" {
		service = "bitcoin"
	}
	if liquid {
		if service == "bitcoin" {
			service = "liquid"
		} else {
			service = fmt.Sprintf("%s-liquid", service)
		}
	}

	return service
}
