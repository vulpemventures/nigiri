package proxy

import (
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	cfg := NewConfig()

	if cfg.IsTLSEnabled() {
		t.Error("TLS should be disabled by default")
	}

	if !cfg.IsFaucetEnabled() {
		t.Error("Faucet should be enabled by default")
	}

	if !cfg.IsMiningEnabled() {
		t.Error("Mining should be enabled by default")
	}

	if cfg.IsLoggerEnabled() {
		t.Error("Logger should be disabled by default")
	}

	if cfg.ListenURL() != "localhost:3000" {
		t.Errorf("Expected listen URL localhost:3000, got %s", cfg.ListenURL())
	}

	if cfg.Chain() != "bitcoin" {
		t.Errorf("Expected chain bitcoin, got %s", cfg.Chain())
	}
}

func TestNewConfig_WithOptions(t *testing.T) {
	cfg := NewConfig(
		WithTLS(true),
		WithFaucet(false),
		WithMining(false),
		WithLogger(true),
		WithListenAddr("0.0.0.0:8080"),
		WithElectrsAddr("electrs:30000"),
		WithRPCAddr("bitcoin", "18443"),
		WithRPCCredentials("user", "pass"),
		WithChain("liquid"),
		WithWalletName("testwallet"),
	)

	if !cfg.IsTLSEnabled() {
		t.Error("TLS should be enabled")
	}

	if cfg.IsFaucetEnabled() {
		t.Error("Faucet should be disabled")
	}

	if cfg.IsMiningEnabled() {
		t.Error("Mining should be disabled")
	}

	if !cfg.IsLoggerEnabled() {
		t.Error("Logger should be enabled")
	}

	if cfg.ListenURL() != "0.0.0.0:8080" {
		t.Errorf("Expected listen URL 0.0.0.0:8080, got %s", cfg.ListenURL())
	}

	if cfg.ElectrsURL() != "http://electrs:30000" {
		t.Errorf("Expected electrs URL http://electrs:30000, got %s", cfg.ElectrsURL())
	}

	if cfg.RPCServerURL() != "http://user:pass@bitcoin:18443" {
		t.Errorf("Expected RPC URL http://user:pass@bitcoin:18443, got %s", cfg.RPCServerURL())
	}

	if cfg.Chain() != "liquid" {
		t.Errorf("Expected chain liquid, got %s", cfg.Chain())
	}

	if cfg.WalletName() != "testwallet" {
		t.Errorf("Expected wallet name testwallet, got %s", cfg.WalletName())
	}
}

func TestNewTestConfig(t *testing.T) {
	cfg := NewTestConfig()

	if cfg.ListenURL() != "localhost:7000" {
		t.Errorf("Expected listen URL localhost:7000, got %s", cfg.ListenURL())
	}

	if cfg.Chain() != "bitcoin" {
		t.Errorf("Expected chain bitcoin, got %s", cfg.Chain())
	}

	if !cfg.IsFaucetEnabled() {
		t.Error("Faucet should be enabled in test config")
	}

	if !cfg.IsMiningEnabled() {
		t.Error("Mining should be enabled in test config")
	}
}

func TestNewLiquidTestConfig(t *testing.T) {
	cfg := NewLiquidTestConfig()

	if cfg.ListenURL() != "localhost:7001" {
		t.Errorf("Expected listen URL localhost:7001, got %s", cfg.ListenURL())
	}

	if cfg.Chain() != "liquid" {
		t.Errorf("Expected chain liquid, got %s", cfg.Chain())
	}
}

func TestWithRegistryPath_AbsolutePath(t *testing.T) {
	cfg := NewConfig(WithRegistryPath("/absolute/path"))

	if cfg.RegistryPath() != "/absolute/path" {
		t.Errorf("Expected registry path /absolute/path, got %s", cfg.RegistryPath())
	}
}

func TestWithRegistryPath_RelativePath(t *testing.T) {
	// Get default config registry path
	defaultCfg := NewConfig()
	defaultPath := defaultCfg.RegistryPath()

	// Relative path should not change the default
	cfg := NewConfig(WithRegistryPath("relative/path"))

	if cfg.RegistryPath() != defaultPath {
		t.Errorf("Relative path should not change registry path, expected %s, got %s", defaultPath, cfg.RegistryPath())
	}
}
