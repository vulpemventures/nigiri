package cmd

import (
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

const (
	btcAddress    = "mpSGWQvbAiRt2UNLST1CdWUufoPVsVwLyK"
	liquidAddress = "CTEsqL1x9ooWWG9HBaHUpvS2DGJJ4haYdkTQPKj9U8CCdwT5vcudhbYUT8oQwwoS11aYtdznobfgT8rj"
)

func TestFaucetBitcoinServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, bitcoin)

	if err := testCommand("faucet", btcAddress, bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestFaucetLiquidServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("faucet", liquidAddress, liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestFaucetShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	expectedError := constants.ErrNigiriNotRunning.Error()

	err := testCommand("faucet", btcAddress, bitcoin)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	err = testCommand("faucet", liquidAddress, liquid)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func TestStartBitcoinAndFaucetNigiriServicesShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, bitcoin)

	expectedError := constants.ErrNigiriLiquidNotEnabled.Error()

	err := testCommand("faucet", liquidAddress, liquid)
	if err == nil {
		t.Fatal("Should return error when trying logging liquid services if not running")
	}

	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	testDelete(t)
}
