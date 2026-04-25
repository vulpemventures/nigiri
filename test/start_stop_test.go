package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/docker"
	"github.com/vulpemventures/nigiri/internal/state"
)

var (
	tmpDatadir       string
	mockClient       *docker.MockClient
	nigiriBinaryPath string // Store path to the built binary
)

func TestMain(m *testing.M) {
	// Determine home directory for temp path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home directory: %v\n", err)
		os.Exit(1)
	}
	// Use a path inside the user's home directory for potentially better Docker mount compatibility
	tmpDatadir = filepath.Join(homeDir, "nigiri-tmp")
	fmt.Printf("Using temp directory: %s\n", tmpDatadir)

	// Ensure the directory does not exist from a previous failed run
	_ = os.RemoveAll(tmpDatadir)

	// Build the nigiri binary for integration tests
	binaryName := "nigiri_test"
	nigiriBinaryPath = filepath.Join(tmpDatadir, binaryName) // Build inside tmpDatadir for easy cleanup

	// Ensure tmpDatadir exists before build
	if err := os.MkdirAll(tmpDatadir, 0755); err != nil {
		fmt.Printf("Failed to create base temp dir for build: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building nigiri binary for tests at %s...\n", nigiriBinaryPath)
	buildCmd := exec.Command("go", "build", "-o", nigiriBinaryPath, "../cmd/nigiri")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build nigiri binary: %v\n", err)
		// Cleanup partially created dir? Best effort.
		_ = os.RemoveAll(tmpDatadir)
		os.Exit(1)
	}
	fmt.Println("Build successful.")

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
    image: elementsproject/lightningd:v25.09.3
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
		name             string
		liquid           bool
		ln               bool
		ark              bool
		expectedServices []string
	}{
		{
			name:             "Basic Services",
			liquid:           false,
			ln:               false,
			ark:              false,
			expectedServices: []string{"bitcoin", "electrs", "chopsticks", "esplora"},
		},
		{
			name:   "With Liquid",
			liquid: true,
			ln:     false,
			ark:    false,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"liquid", "electrs-liquid", "chopsticks-liquid", "esplora-liquid",
			},
		},
		{
			name:   "With Lightning",
			liquid: false,
			ln:     true,
			ark:    false,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"lnd", "tap", "cln",
			},
		},
		{
			name:   "With Ark",
			liquid: false,
			ln:     false,
			ark:    true,
			expectedServices: []string{
				"bitcoin", "electrs", "chopsticks", "esplora",
				"ark",
			},
		},
		{
			name:   "All Services",
			liquid: true,
			ln:     true,
			ark:    true,
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

// Helper to run the built nigiri binary
func runNigiri(t *testing.T, args ...string) error {
	t.Helper()
	cmdArgs := append([]string{"--datadir", tmpDatadir}, args...)
	cmd := exec.Command(nigiriBinaryPath, cmdArgs...)
	cmd.Stdout = os.Stdout // Or capture if needed
	cmd.Stderr = os.Stderr
	t.Logf("Running: %s %v", nigiriBinaryPath, cmdArgs)
	err := cmd.Run()
	if err != nil {
		t.Logf("Command failed: %v", err)
	}
	return err // Return error for checking
}

// Helper to check flags file content (semantic comparison)
func checkFlagsFile(t *testing.T, expectExists bool, expectedJsonString string) {
	t.Helper()
	flagsFilePath := filepath.Join(tmpDatadir, "flags.json") // Assuming name defined in start.go
	_, err := os.Stat(flagsFilePath)

	if expectExists {
		if os.IsNotExist(err) {
			t.Fatalf("Expected flags file '%s' to exist, but it doesn't", flagsFilePath)
		}
		if err != nil {
			t.Fatalf("Error checking flags file '%s': %v", flagsFilePath, err)
		}
		// Read actual content
		actualContentBytes, readErr := os.ReadFile(flagsFilePath)
		if readErr != nil {
			t.Fatalf("Failed to read flags file '%s': %v", flagsFilePath, readErr)
		}

		// Unmarshal both expected and actual into maps for comparison
		var expectedMap, actualMap map[string]interface{}

		errUnmarshalExpected := json.Unmarshal([]byte(expectedJsonString), &expectedMap)
		if errUnmarshalExpected != nil {
			t.Fatalf("Failed to unmarshal expected JSON string: %v\nString was:\n%s", errUnmarshalExpected, expectedJsonString)
		}

		errUnmarshalActual := json.Unmarshal(actualContentBytes, &actualMap)
		if errUnmarshalActual != nil {
			t.Fatalf("Failed to unmarshal actual flags file content: %v\nContent was:\n%s", errUnmarshalActual, string(actualContentBytes))
		}

		// Compare maps using reflect.DeepEqual
		if !reflect.DeepEqual(expectedMap, actualMap) {
			t.Errorf("Flags file content mismatch (semantic).\nExpected: %v\nGot:      %v", expectedMap, actualMap)
		}

	} else {
		if err == nil {
			t.Fatalf("Expected flags file '%s' to not exist, but it does", flagsFilePath)
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Error checking non-existence of flags file '%s': %v", flagsFilePath, err)
		}
	}
}

func TestRememberForget(t *testing.T) {
	// Ensure state is clean before starting
	stateFilePath := filepath.Join(tmpDatadir, config.DefaultName) // Use config constant
	flagsFilePath := filepath.Join(tmpDatadir, "flags.json")       // Matches start.go

	errState := os.Remove(stateFilePath)
	// Ignore 'not exist' error, but fail on others
	if errState != nil && !os.IsNotExist(errState) {
		t.Fatalf("Failed to remove previous state file '%s' before test: %v", stateFilePath, errState)
	}
	errFlags := os.Remove(flagsFilePath)
	// Ignore 'not exist' error, but fail on others
	if errFlags != nil && !os.IsNotExist(errFlags) {
		t.Fatalf("Failed to remove previous flags file '%s' before test: %v", flagsFilePath, errFlags)
	}

	// 1. Start with --liquid and --remember
	t.Log("Running: start --liquid --remember")
	if err := runNigiri(t, "start", "--liquid", "--remember"); err != nil {
		// Start might fail if docker-compose isn't mocked/available,
		// but we primarily care about the flags file for this test.
		// Let's only log the error for now, but check the file.
		t.Logf("Note: 'start' command returned error (expected in CI without docker?): %v", err)
	}
	// Check flags.json was created with liquid: true
	// Need to marshal the expected struct to JSON string for comparison
	expectedFlags := map[string]bool{"liquid": true, "ln": false, "ark": false, "ci": false}
	expectedJsonBytes, _ := json.MarshalIndent(expectedFlags, "", "  ")
	checkFlagsFile(t, true, string(expectedJsonBytes))

	// 2. Stop
	t.Log("Running: stop")
	_ = runNigiri(t, "stop") // Ignore error for now

	// 3. Start again (should load remembered flags)
	t.Log("Running: start (expecting remembered flags)")
	// We can't easily verify docker services without mocking docker-compose,
	// but we can ensure the flags file *wasn't* overwritten if --remember wasn't used.
	_ = runNigiri(t, "start")
	checkFlagsFile(t, true, string(expectedJsonBytes)) // Should still contain the remembered flags

	// 4. Stop again
	t.Log("Running: stop")
	_ = runNigiri(t, "stop")

	// 5. Forget
	t.Log("Running: forget")
	if err := runNigiri(t, "forget"); err != nil {
		t.Fatalf("forget command failed: %v", err)
	}
	// Check flags.json was removed
	checkFlagsFile(t, false, "")

	// 6. Start again (should use defaults, no liquid)
	t.Log("Running: start (expecting default flags)")
	_ = runNigiri(t, "start")
	// Check flags.json does *not* exist (because --remember wasn't used)
	checkFlagsFile(t, false, "")

	// 7. Stop finally
	t.Log("Running: stop")
	_ = runNigiri(t, "stop")
}

func TestUpdateForce(t *testing.T) {
	// Create a mock update script
	mockScript := `#!/bin/bash
echo "Mock update script executed successfully"
exit 0
`

	// Create a test HTTP server that serves the mock script
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockScript))
	}))
	defer server.Close()

	// Set the environment variable to use the test server URL
	originalURL := os.Getenv("NIGIRI_TEST_UPDATE_SCRIPT_URL")
	os.Setenv("NIGIRI_TEST_UPDATE_SCRIPT_URL", server.URL)
	defer func() {
		if originalURL == "" {
			os.Unsetenv("NIGIRI_TEST_UPDATE_SCRIPT_URL")
		} else {
			os.Setenv("NIGIRI_TEST_UPDATE_SCRIPT_URL", originalURL)
		}
	}()

	// Run the update --force command
	// Note: This will replace the current process with the mock script,
	// so we need to run it in a separate process
	cmd := exec.Command(nigiriBinaryPath, "--datadir", tmpDatadir, "update", "--force")
	cmd.Env = append(os.Environ(), "NIGIRI_TEST_UPDATE_SCRIPT_URL="+server.URL)
	
	output, err := cmd.CombinedOutput()
	
	// The command should execute the mock script successfully
	// Since syscall.Exec replaces the process, we expect the output to be from the mock script
	if err != nil {
		t.Logf("Command output: %s", string(output))
		// Check if the output contains our expected message
		if !strings.Contains(string(output), "Mock update script executed successfully") {
			t.Fatalf("update --force failed: %v\nOutput: %s", err, string(output))
		}
	}
	
	// Verify that the download message was printed
	if !strings.Contains(string(output), "Downloading update script") {
		t.Errorf("Expected download message in output, got: %s", string(output))
	}
}

func TestCustomCompose(t *testing.T) {
	// Clean up any previous state
	stateFilePath := filepath.Join(tmpDatadir, config.DefaultName)
	flagsFilePath := filepath.Join(tmpDatadir, "flags.json")
	_ = os.Remove(stateFilePath)
	_ = os.Remove(flagsFilePath)

	// Create a custom docker-compose file
	customComposePath := filepath.Join(tmpDatadir, "custom-compose.yml")
	customComposeContent := `name: nigiri-custom
services:
  bitcoin:
    image: getumbrel/bitcoind:v28.0
    container_name: bitcoin-custom
  electrs:
    image: vulpemventures/electrs:latest
  chopsticks:
    image: vulpemventures/nigiri-chopsticks:latest
  esplora:
    image: vulpemventures/esplora:latest
`
	
	if err := os.WriteFile(customComposePath, []byte(customComposeContent), 0644); err != nil {
		t.Fatalf("Failed to create custom compose file: %v", err)
	}
	defer os.Remove(customComposePath)

	// Test 1: Start with custom compose and --remember
	t.Log("Running: start --compose custom-compose.yml --remember")
	cmd := exec.Command(nigiriBinaryPath, "--datadir", tmpDatadir, "start", "--compose", customComposePath, "--remember")
	output, err := cmd.CombinedOutput()
	
	// We expect it to fail because docker-compose isn't available in CI,
	// but it should at least accept the flag and find the file
	if err != nil {
		// Check that it's not failing due to file not found
		if strings.Contains(string(output), "custom compose file not found") {
			t.Fatalf("Custom compose file should have been found: %s", string(output))
		}
		// Expected error is docker-compose not found, which is fine
		t.Logf("Expected docker-compose error (first start): %s", string(output))
	}

	// Verify flags.json was created with the compose path
	expectedFlags := map[string]interface{}{
		"liquid":  false,
		"ln":      false,
		"ark":     false,
		"ci":      false,
		"compose": customComposePath,
	}
	expectedJsonBytes, _ := json.MarshalIndent(expectedFlags, "", "  ")
	checkFlagsFile(t, true, string(expectedJsonBytes))

	// Stop
	t.Log("Running: stop")
	_ = runNigiri(t, "stop")

	// Test 2: Start again without --compose flag (should use remembered path)
	t.Log("Running: start (expecting remembered compose path)")
	cmd = exec.Command(nigiriBinaryPath, "--datadir", tmpDatadir, "start")
	output, err = cmd.CombinedOutput()
	
	if err != nil {
		if strings.Contains(string(output), "custom compose file not found") {
			t.Fatalf("Should have used remembered compose path: %s", string(output))
		}
		t.Logf("Expected docker-compose error (second start): %s", string(output))
	}

	// Flags should still contain the compose path
	checkFlagsFile(t, true, string(expectedJsonBytes))

	// Stop
	t.Log("Running: stop")
	_ = runNigiri(t, "stop")

	// Test 3: Test with non-existent file
	cmd = exec.Command(nigiriBinaryPath, "--datadir", tmpDatadir, "start", "--compose", "/nonexistent/compose.yml")
	output, err = cmd.CombinedOutput()
	
	if err == nil {
		t.Fatal("Expected error for non-existent compose file")
	}
	
	if !strings.Contains(string(output), "custom compose file not found") {
		t.Errorf("Expected 'custom compose file not found' error, got: %s", string(output))
	}

	// Cleanup
	_ = runNigiri(t, "forget")
}
