package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

func TestRpcBitcoinCommand(t *testing.T) {
	testStart(t, bitcoin)

	bashCmd := exec.Command("docker", "ps", "-a")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr
	bashCmd.Run()

	if err := testCommand("rpc", "getblockchaininfo", bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcBitcoinTwoCommands(t *testing.T) {
	testStart(t, bitcoin)

	if err := testCommand("rpc", "getblockhash 0", bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcLiquidCommand(t *testing.T) {
	testStart(t, liquid)

	if err := testCommand("rpc", "getblockchaininfo", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcLiquidTwoCommands(t *testing.T) {
	testStart(t, liquid)

	if err := testCommand("rpc", "getblockhash 0", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcShouldFail(t *testing.T) {
	expectedError := constants.ErrNigiriNotRunning.Error()

	err := testCommand("rpc", "getblockchaininfo", bitcoin)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	err = testCommand("rpc", "getblockchaininfo", liquid)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func TestStartBitcoinAndRpcNigiriServicesShouldFail(t *testing.T) {
	testStart(t, bitcoin)

	expectedError := constants.ErrNigiriLiquidNotEnabled.Error()

	err := testCommand("rpc", "getblockchaininfo", liquid)
	if err == nil {
		t.Fatal("Should return error when trying logging liquid services if not running")
	}

	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	testDelete(t)
}
