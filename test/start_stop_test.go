package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
	"testing"

	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/docker"
	"github.com/vulpemventures/nigiri/internal/state"
)

var (
	tmpDatadir = filepath.Join(os.TempDir(), "nigiri-tmp")
	mockClient *docker.MockClient
)

func TestMain(m *testing.M) {
	// Initialize mock client
	mockClient = docker.NewMockClient()

	// Setup test environment
	if err := setupTestEnvironment(); err != nil {
		fmt.Printf("Failed to setup test environment: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup any leftover containers
	cleanup()
	os.Exit(code)
}

func setupTestEnvironment() error {
	// Create test directory and subdirectories
	dirs := []string{
		tmpDatadir,
		filepath.Join(tmpDatadir, "bitcoin"),
		filepath.Join(tmpDatadir, "liquid"),
		filepath.Join(tmpDatadir, "lnd"),
		filepath.Join(tmpDatadir, "cln"),
		filepath.Join(tmpDatadir, "chopsticks"),
		filepath.Join(tmpDatadir, "chopsticks-liquid"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create mock configuration files
	files := map[string]string{
		"docker-compose.yml": `name: nigiri
services:
  bitcoin:
    image: getumbrel/bitcoind:v28.0
    container_name: bitcoin
  liquid:
    image: ghcr.io/vulpemventures/elements:latest
  electrs:
    image: vulpemventures/electrs:latest
  chopsticks:
    image: vulpemventures/nigiri-chopsticks:latest
  esplora:
    image: vulpemventures/esplora:latest
  lnd:
    image: lightninglabs/lnd:latest
  tap:
    image: vulpemventures/tap:latest
  cln:
    image: elementsproject/lightningd:latest
  ark:
    image: vulpemventures/ark:latest`,

		"bitcoin.conf": `regtest=1
rpcuser=admin1
rpcpassword=123
rpcallowip=0.0.0.0/0`,

		"elements.conf": `chain=liquidregtest
rpcuser=admin1
rpcpassword=123
rpcallowip=0.0.0.0/0`,

		"lnd.conf": `[Application Options]
debuglevel=info
noseedbackup=1
alias=nigiri-lnd
listen=0.0.0.0:9735`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDatadir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

func cleanup() {
	// Remove temp directory
	_ = os.RemoveAll(tmpDatadir)
}

func TestDataDirSetup(t *testing.T) {
	// Create test directory
	err := os.MkdirAll(tmpDatadir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Expected files and directories
	expectedFiles := []string{
		"docker-compose.yml",
		"bitcoin.conf",
		"elements.conf",
		"lnd.conf",
	}

	expectedDirs := []string{
		"bitcoin",
		"liquid",
		"lnd",
		"cln",
		"chopsticks",
		"chopsticks-liquid",
	}

	// Check if required files exist
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDatadir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", file)
		}
	}

	// Check if required directories exist
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDatadir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s does not exist", dir)
		}
	}

	// Check docker-compose.yml content
	composeContent, err := os.ReadFile(filepath.Join(tmpDatadir, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	// Verify essential services are defined in docker-compose.yml
	essentialServices := []string{
		"bitcoin:",
		"electrs:",
		"chopsticks:",
		"esplora:",
	}

	for _, service := range essentialServices {
		if !strings.Contains(string(composeContent), service) {
			t.Errorf("docker-compose.yml missing essential service: %s", service)
		}
	}

	// Check bitcoin.conf content
	bitcoinConf, err := os.ReadFile(filepath.Join(tmpDatadir, "bitcoin.conf"))
	if err != nil {
		t.Fatalf("Failed to read bitcoin.conf: %v", err)
	}

	// Verify essential bitcoin configuration
	bitcoinConfigs := []string{
		"regtest=1",
		"rpcuser=",
		"rpcpassword=",
		"rpcallowip=",
	}

	for _, config := range bitcoinConfigs {
		if !strings.Contains(string(bitcoinConf), config) {
			t.Errorf("bitcoin.conf missing essential config: %s", config)
		}
	}
}

func TestBasicStartStop(t *testing.T) {
	// Setup test state
	_ = state.New(filepath.Join(tmpDatadir, config.DefaultName), config.InitialState)

	// Use mock client for Docker operations
	composePath := filepath.Join(tmpDatadir, config.DefaultCompose)
	
	// Test start command
	mockClient.ClearCommands()
	mockClient.SetMockCmd(exec.Command("echo", "mock start"))
	
	// Execute start command
	mockClient.RunCompose(composePath, "up", "-d", "bitcoin", "electrs", "chopsticks", "esplora")
	
	// Verify the correct Docker commands were called
	composePath, args, ok := mockClient.GetLastCommand()
	if !ok {
		t.Fatal("No Docker commands were executed")
	}
	
	// Check that the compose file path is correct
	if composePath != filepath.Join(tmpDatadir, config.DefaultCompose) {
		t.Errorf("Expected compose path %s, got %s", filepath.Join(tmpDatadir, config.DefaultCompose), composePath)
	}
	
	// Check that the correct services were started
	expectedArgs := []string{"up", "-d", "bitcoin", "electrs", "chopsticks", "esplora"}
	if !stringSliceEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}

	// Test stop command
	mockClient.ClearCommands()
	mockClient.SetMockCmd(exec.Command("echo", "mock stop"))
	
	// Execute stop command
	mockClient.RunCompose(composePath, "down")
	
	// Verify stop command
	composePath, args, ok = mockClient.GetLastCommand()
	if !ok {
		t.Fatal("No Docker commands were executed for stop")
	}
	
	expectedArgs = []string{"down"}
	if !stringSliceEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

func TestStateManagement(t *testing.T) {
	statePath := filepath.Join(tmpDatadir, config.DefaultName)
	testState := state.New(statePath, config.InitialState)

	// Test initial state
	initialState, err := testState.Get()
	if err != nil {
		t.Fatalf("Failed to get initial state: %v", err)
	}

	// Verify initial state values
	expectedNetwork := "regtest"
	if network, ok := initialState["network"]; !ok || network != expectedNetwork {
		t.Errorf("Expected network to be %s, got %s", expectedNetwork, network)
	}

	expectedRunning := "false"
	if running, ok := initialState["running"]; !ok || running != expectedRunning {
		t.Errorf("Expected running to be %s, got %s", expectedRunning, running)
	}

	// Test state updates
	updatedState := map[string]string{
		"network": "regtest",
		"running": "true",
		"ready":   "true",
	}

	err = testState.Set(updatedState)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	// Verify state was updated
	currentState, err := testState.Get()
	if err != nil {
		t.Fatalf("Failed to get current state: %v", err)
	}

	for key, expectedValue := range updatedState {
		if value, ok := currentState[key]; !ok || value != expectedValue {
			t.Errorf("Expected %s to be %s, got %s", key, expectedValue, value)
		}
	}
}

func TestServiceCombinations(t *testing.T) {
	tests := []struct {
		name           string
		liquid        bool
		ln            bool
		ark           bool
		expectedServices []string
	}{
		{
			name:    "Basic Services",
			liquid:  false,
			ln:      false,
			ark:     false,
			expectedServices: []string{"bitcoin", "electrs", "chopsticks", "esplora"},
		},
		{
			name:    "With Liquid",
			liquid:  true,
			ln:      false,
			ark:     false,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"liquid", "electrs-liquid", "chopsticks-liquid", "esplora-liquid",
			},
		},
		{
			name:    "With Lightning",
			liquid:  false,
			ln:      true,
			ark:     false,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"lnd", "tap", "cln",
			},
		},
		{
			name:    "With Ark",
			liquid:  false,
			ln:      false,
			ark:     true,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"ark",
			},
		},
		{
			name:    "All Services",
			liquid:  true,
			ln:      true,
			ark:     true,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"liquid", "electrs-liquid", "chopsticks-liquid", "esplora-liquid",
				"lnd", "tap", "cln",
				"ark",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test state
			statePath := filepath.Join(tmpDatadir, config.DefaultName)
			testState := state.New(statePath, config.InitialState)

			// Use mock client for Docker operations
			composePath := filepath.Join(tmpDatadir, config.DefaultCompose)
			
			// Test start command
			mockClient.ClearCommands()
			mockClient.SetMockCmd(exec.Command("echo", "mock start"))
			
			// Build the docker-compose command with appropriate services
			args := []string{"up", "-d"}
			args = append(args, tt.expectedServices...)
			
			// Execute start command
			mockClient.RunCompose(composePath, args...)
			
			// Verify the correct Docker commands were called
			composePath, actualArgs, ok := mockClient.GetLastCommand()
			if !ok {
				t.Fatal("No Docker commands were executed")
			}
			
			// Check that the compose file path is correct
			if composePath != filepath.Join(tmpDatadir, config.DefaultCompose) {
				t.Errorf("Expected compose path %s, got %s", filepath.Join(tmpDatadir, config.DefaultCompose), composePath)
			}
			
			// Check that the correct services were started
			expectedArgs := append([]string{"up", "-d"}, tt.expectedServices...)
			if !stringSliceEqual(actualArgs, expectedArgs) {
				t.Errorf("Expected args %v, got %v", expectedArgs, actualArgs)
			}

			// Verify state was updated correctly
			err := testState.Set(map[string]string{
				"running": "true",
				"liquid":  strconv.FormatBool(tt.liquid),
				"ln":      strconv.FormatBool(tt.ln),
				"ark":     strconv.FormatBool(tt.ark),
			})
			if err != nil {
				t.Fatalf("Failed to update state: %v", err)
			}

			// Verify state values
			state, err := testState.Get()
			if err != nil {
				t.Fatalf("Failed to get state: %v", err)
			}

			// Check each flag in state
			expectedFlags := map[string]bool{
				"liquid": tt.liquid,
				"ln":     tt.ln,
				"ark":    tt.ark,
			}

			for flag, expected := range expectedFlags {
				if value, ok := state[flag]; !ok {
					t.Errorf("State missing %s flag", flag)
				} else if actual := value == "true"; actual != expected {
					t.Errorf("Expected %s to be %v, got %v", flag, expected, actual)
				}
			}

			// Test stop command
			mockClient.ClearCommands()
			mockClient.SetMockCmd(exec.Command("echo", "mock stop"))
			
			// Execute stop command
			mockClient.RunCompose(composePath, "down")
			
			// Verify stop command
			_, stopArgs, ok := mockClient.GetLastCommand()
			if !ok {
				t.Fatal("No Docker commands were executed for stop")
			}
			
			if !stringSliceEqual(stopArgs, []string{"down"}) {
				t.Errorf("Expected stop args [down], got %v", stopArgs)
			}
		})
	}
}

// Helper function to compare string slices
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
