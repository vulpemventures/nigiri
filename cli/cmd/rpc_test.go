package cmd

import (
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

func TestRpcBitcoinCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, bitcoin)

	if err := testCommand("rpc", "getblockchaininfo", bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcBitcoinTwoCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, bitcoin)

	if err := testCommand("rpc", "getblockhash 0", bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcLiquidCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("rpc", "getblockchaininfo", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcLiquidTwoCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("rpc", "getblockhash 0", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestRpcShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
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
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
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
