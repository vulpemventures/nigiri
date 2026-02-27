// +build e2e

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// E2E Smoke Test for Nigiri with embedded chopsticks proxy
// Run with: go test -tags=e2e -v ./test -run TestE2ESmoke
//
// Prerequisites:
// - Docker running (OrbStack, Docker Desktop, etc.)
// - nigiri binary built: go build -o nigiri ./cmd/nigiri
// - No existing nigiri instance running

const (
	proxyURL     = "http://localhost:3000"
	startTimeout = 120 * time.Second
	pollInterval = 2 * time.Second
)

func TestE2ESmoke(t *testing.T) {
	// Skip if not in e2e mode
	if os.Getenv("NIGIRI_E2E") != "1" && !isE2EBuild() {
		t.Skip("Skipping e2e test. Run with -tags=e2e or set NIGIRI_E2E=1")
	}

	// Ensure clean state
	cleanup(t)
	defer cleanup(t)

	t.Run("StartNigiri", testStartNigiri)
	t.Run("WaitForServices", testWaitForServices)
	t.Run("ElectrsProxy", testElectrsProxy)
	t.Run("Faucet", testFaucet)
	t.Run("BlockchainState", testBlockchainState)
}

func isE2EBuild() bool {
	// This function returns true when built with -tags=e2e
	// due to the build constraint at the top of the file
	return true
}

func cleanup(t *testing.T) {
	t.Helper()
	cmd := exec.Command("./nigiri", "stop", "--delete")
	cmd.Run() // Ignore errors, might not be running
	time.Sleep(2 * time.Second)
}

func testStartNigiri(t *testing.T) {
	cmd := exec.Command("./nigiri", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to start nigiri: %v", err)
	}

	t.Log("✓ nigiri start completed")
}

func testWaitForServices(t *testing.T) {
	deadline := time.Now().Add(startTimeout)

	// Wait for proxy to be ready
	for time.Now().Before(deadline) {
		resp, err := http.Get(proxyURL + "/blocks/tip/height")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Log("✓ Proxy is ready")
				return
			}
		}
		time.Sleep(pollInterval)
	}

	t.Fatal("Timeout waiting for services to be ready")
}

func testElectrsProxy(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		method   string
		wantCode int
	}{
		{"BlockHeight", "/blocks/tip/height", "GET", http.StatusOK},
		{"BlockHash", "/blocks/tip/hash", "GET", http.StatusOK},
		{"FeeEstimates", "/fee-estimates", "GET", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(proxyURL + tt.endpoint)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantCode {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("GET %s: got status %d, want %d. Body: %s",
					tt.endpoint, resp.StatusCode, tt.wantCode, string(body))
			} else {
				t.Logf("✓ GET %s returned %d", tt.endpoint, resp.StatusCode)
			}
		})
	}
}

func testFaucet(t *testing.T) {
	// First, get a new address from bitcoind
	address := getNewAddress(t)
	if address == "" {
		t.Fatal("Failed to get new address")
	}
	t.Logf("✓ Got new address: %s", address)

	// Send faucet request
	faucetReq := map[string]interface{}{
		"address": address,
		"amount":  1.0,
	}
	reqBody, _ := json.Marshal(faucetReq)

	resp, err := http.Post(proxyURL+"/faucet", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Faucet request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Faucet returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse faucet response: %v", err)
	}

	txid := result["txId"]
	if txid == "" {
		t.Fatalf("Faucet response missing txId: %s", string(body))
	}

	t.Logf("✓ Faucet sent transaction: %s", txid)

	// Verify transaction exists
	time.Sleep(2 * time.Second) // Wait for electrs to index

	txResp, err := http.Get(proxyURL + "/tx/" + txid)
	if err != nil {
		t.Fatalf("Failed to fetch transaction: %v", err)
	}
	defer txResp.Body.Close()

	if txResp.StatusCode != http.StatusOK {
		t.Errorf("Transaction %s not found via electrs proxy", txid)
	} else {
		t.Logf("✓ Transaction confirmed and indexed")
	}
}

func testBlockchainState(t *testing.T) {
	// Check block height > 0 (should have mined blocks for faucet)
	resp, err := http.Get(proxyURL + "/blocks/tip/height")
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	height := strings.TrimSpace(string(body))

	if height == "0" {
		t.Error("Block height is still 0, expected blocks to be mined")
	} else {
		t.Logf("✓ Current block height: %s", height)
	}
}

func getNewAddress(t *testing.T) string {
	t.Helper()

	// Try via proxy endpoint first
	resp, err := http.Get(proxyURL + "/getnewaddress")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var result map[string]string
			if json.Unmarshal(body, &result) == nil {
				if addr := result["address"]; addr != "" {
					return addr
				}
			}
		}
	}

	// Fallback to docker exec
	cmd := exec.Command("docker", "exec", "bitcoin",
		"bitcoin-cli", "-regtest", "-rpcuser=admin1", "-rpcpassword=123", "getnewaddress")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Warning: failed to get address via docker: %v", err)
		return ""
	}

	return strings.TrimSpace(string(output))
}

// TestE2ESmokeQuick is a faster version that just checks connectivity
func TestE2ESmokeQuick(t *testing.T) {
	if os.Getenv("NIGIRI_E2E_QUICK") != "1" {
		t.Skip("Skipping quick e2e test. Set NIGIRI_E2E_QUICK=1 to run")
	}

	// Just check if proxy is responding (assumes nigiri is already running)
	resp, err := http.Get(proxyURL + "/blocks/tip/height")
	if err != nil {
		t.Fatalf("Proxy not responding: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Proxy returned %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Block height: %s\n", strings.TrimSpace(string(body)))
	t.Log("✓ Proxy is healthy")
}
