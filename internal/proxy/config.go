package proxy

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the configuration for the proxy server
type Config interface {
	IsTLSEnabled() bool
	IsFaucetEnabled() bool
	IsLoggerEnabled() bool
	IsMiningEnabled() bool
	ListenURL() string
	RPCServerURL() string
	ElectrsURL() string
	Chain() string
	RegistryPath() string
	WalletName() string
}

type config struct {
	tlsEnabled    bool
	faucetEnabled bool
	miningEnabled bool
	loggerEnabled bool
	listenAddr    string
	electrsAddr   string
	rpcUser       string
	rpcPassword   string
	rpcHost       string
	rpcPort       string
	chain         string
	registryPath  string
	walletName    string
}

// ConfigOption is a functional option for configuring the proxy
type ConfigOption func(*config)

// NewConfig creates a new Config with the given options
func NewConfig(opts ...ConfigOption) Config {
	c := &config{
		tlsEnabled:    false,
		faucetEnabled: true,
		miningEnabled: true,
		loggerEnabled: false,
		listenAddr:    "localhost:3000",
		electrsAddr:   "localhost:30000",
		rpcUser:       "admin1",
		rpcPassword:   "123",
		rpcHost:       "localhost",
		rpcPort:       "18443",
		chain:         "bitcoin",
		walletName:    "",
	}

	// Set default registry path
	cwd, _ := os.Getwd()
	c.registryPath = cwd

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithTLS enables TLS
func WithTLS(enabled bool) ConfigOption {
	return func(c *config) {
		c.tlsEnabled = enabled
	}
}

// WithFaucet enables/disables the faucet
func WithFaucet(enabled bool) ConfigOption {
	return func(c *config) {
		c.faucetEnabled = enabled
	}
}

// WithMining enables/disables mining after broadcasts
func WithMining(enabled bool) ConfigOption {
	return func(c *config) {
		c.miningEnabled = enabled
	}
}

// WithLogger enables/disables request logging
func WithLogger(enabled bool) ConfigOption {
	return func(c *config) {
		c.loggerEnabled = enabled
	}
}

// WithListenAddr sets the listen address
func WithListenAddr(addr string) ConfigOption {
	return func(c *config) {
		c.listenAddr = addr
	}
}

// WithElectrsAddr sets the electrs address
func WithElectrsAddr(addr string) ConfigOption {
	return func(c *config) {
		c.electrsAddr = addr
	}
}

// WithRPCCredentials sets the RPC credentials
func WithRPCCredentials(user, password string) ConfigOption {
	return func(c *config) {
		c.rpcUser = user
		c.rpcPassword = password
	}
}

// WithRPCAddr sets the RPC server address
func WithRPCAddr(host, port string) ConfigOption {
	return func(c *config) {
		c.rpcHost = host
		c.rpcPort = port
	}
}

// WithChain sets the chain type (bitcoin or liquid)
func WithChain(chain string) ConfigOption {
	return func(c *config) {
		c.chain = chain
	}
}

// WithRegistryPath sets the registry path for Liquid assets
func WithRegistryPath(path string) ConfigOption {
	return func(c *config) {
		if filepath.IsAbs(path) {
			c.registryPath = path
		}
	}
}

// WithWalletName sets the wallet name
func WithWalletName(name string) ConfigOption {
	return func(c *config) {
		c.walletName = name
	}
}

func (c *config) IsTLSEnabled() bool {
	return c.tlsEnabled
}

func (c *config) IsFaucetEnabled() bool {
	return c.faucetEnabled
}

func (c *config) IsLoggerEnabled() bool {
	return c.loggerEnabled
}

func (c *config) IsMiningEnabled() bool {
	return c.miningEnabled
}

func (c *config) ListenURL() string {
	return c.listenAddr
}

func (c *config) RPCServerURL() string {
	return fmt.Sprintf("http://%s:%s@%s:%s", c.rpcUser, c.rpcPassword, c.rpcHost, c.rpcPort)
}

func (c *config) ElectrsURL() string {
	return fmt.Sprintf("http://%s", c.electrsAddr)
}

func (c *config) Chain() string {
	return c.chain
}

func (c *config) WalletName() string {
	return c.walletName
}

func (c *config) RegistryPath() string {
	return c.registryPath
}

// NewTestConfig returns a config suitable for testing
func NewTestConfig() Config {
	return NewConfig(
		WithListenAddr("localhost:7000"),
		WithElectrsAddr("localhost:30000"),
		WithRPCAddr("localhost", "18443"),
		WithRPCCredentials("admin1", "123"),
		WithFaucet(true),
		WithMining(true),
		WithChain("bitcoin"),
	)
}

// NewLiquidTestConfig returns a config suitable for Liquid testing
func NewLiquidTestConfig() Config {
	cwd, _ := os.Getwd()
	return NewConfig(
		WithListenAddr("localhost:7001"),
		WithElectrsAddr("localhost:30001"),
		WithRPCAddr("localhost", "18884"),
		WithRPCCredentials("admin1", "123"),
		WithFaucet(true),
		WithMining(true),
		WithChain("liquid"),
		WithRegistryPath(cwd),
	)
}
