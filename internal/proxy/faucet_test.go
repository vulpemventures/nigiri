package proxy

import (
	"testing"
)

func TestNewFaucet(t *testing.T) {
	// Test that NewFaucet creates a valid Faucet instance
	client, _ := NewRPCClient("http://localhost:18443", false, 10)
	faucet := NewFaucet("http://localhost:18443", client)

	if faucet == nil {
		t.Error("Expected non-nil Faucet")
	}

	if faucet.URL != "http://localhost:18443" {
		t.Errorf("Expected URL http://localhost:18443, got %s", faucet.URL)
	}

	if faucet.rpcClient != client {
		t.Error("Expected rpcClient to be set")
	}
}
