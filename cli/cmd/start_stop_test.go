package cmd

import (
	"fmt"
	"testing"
)

const (
	liquid  = true
	bitcoin = false
	delete  = true
)

var (
	stopCmd = []string{"stop"}
	// deleteCmd      = append(stopCmd, "--delete")
	startCmd = []string{"start"}
	// liquidStartCmd = append(startCmd, "--liquid")
)

func TestStartStopLiquid(t *testing.T) {
	// Start/Stop
	testStart(t, liquid)
	testStop(t)
	// Start/Delete
	testStart(t, liquid)
	testDelete(t)
}

func TestStartStopBitcoin(t *testing.T) {
	// Start/Stop
	testStart(t, bitcoin)
	testStop(t)
	// Start/Delete
	testStart(t, bitcoin)
	testDelete(t)
}

func TestStopBeforeStartShouldFail(t *testing.T) {
	expectedError := "Nigiri is neither running nor stopped, please create it first"

	err := testCommand("stop", "", !delete)
	if err == nil {
		t.Fatal("Should return error when trying to stop before starting")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	err = testCommand("stop", "", delete)
	if err == nil {
		t.Fatal("Should return error when trying to delete before starting")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func TestStartAfterStartShouldFail(t *testing.T) {
	expectedError := "Nigiri is already running, please stop it first"

	if err := testCommand("start", "", bitcoin); err != nil {
		t.Fatal(err)
	}

	err := testCommand("start", "", bitcoin)
	if err == nil {
		t.Fatal("Should return error when trying to start Nigiri if already started")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	err = testCommand("start", "", liquid)
	if err == nil {
		t.Fatal("Should return error when trying to start Nigiri if already started")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func testStart(t *testing.T, flag bool) {
	if err := testCommand("start", "", flag); err != nil {
		t.Fatal(err)
	}
	if isRunning, _ := nigiriIsRunning(); !isRunning {
		t.Fatal("Nigiri should be started but services have not been found among running containers")
	}
}

func testStop(t *testing.T) {
	fmt.Println(!delete)
	if err := testCommand("stop", "", !delete); err != nil {
		t.Fatal(err)
	}
	if isStopped, _ := nigiriExistsAndNotRunning(); !isStopped {
		t.Fatal("Nigiri should be stopped but services have not been found among stopped containers")
	}
}

func testDelete(t *testing.T) {
	if err := testCommand("stop", "", delete); err != nil {
		t.Fatal(err)
	}
	if isStopped, _ := nigiriExistsAndNotRunning(); isStopped {
		t.Fatal("Nigiri should be terminated at this point but services have found among stopped containers")
	}
}

func testCommand(command, arg string, flag bool) error {
	cmd := RootCmd
	cmd.SetArgs(nil)

	if command == "start" {
		args := append(startCmd, fmt.Sprintf("--liquid=%t", flag))
		cmd.SetArgs(args)
	}
	if command == "stop" {
		args := append(stopCmd, fmt.Sprintf("--delete=%t", flag))
		cmd.SetArgs(args)
	}
	if command == "logs" {
		logsCmd := []string{command, arg, fmt.Sprintf("--liquid=%t", flag)}
		cmd.SetArgs(logsCmd)
	}

	if err := cmd.Execute(); err != nil {
		return err
	}

	return nil
}
