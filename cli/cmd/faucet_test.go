package cmd

import (
	"testing"
	"time"

	"github.com/vulpemventures/nigiri/cli/constants"
)

const (
	btcAddress    = "mpSGWQvbAiRt2UNLST1CdWUufoPVsVwLyK"
	liquidAddress = "CTEsqL1x9ooWWG9HBaHUpvS2DGJJ4haYdkTQPKj9U8CCdwT5vcudhbYUT8oQwwoS11aYtdznobfgT8rj"
)

func TestFaucetBitcoinServices(t *testing.T) {
	testStart(t, bitcoin)

	//Give some time to nigiri to be ready before calling
	time.Sleep(2 * time.Second)

	if err := testCommand("faucet", btcAddress, bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestFaucetLiquidServices(t *testing.T) {
	testStart(t, liquid)

	//Give some time to nigiri to be ready before calling
	time.Sleep(2 * time.Second)

	if err := testCommand("faucet", liquidAddress, liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestFaucetShouldFail(t *testing.T) {
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
	if err := testCommand("start", "", bitcoin); err != nil {
		t.Fatal(err)
	}

	expectedError := constants.ErrNigiriLiquidNotEnabled.Error()

	err := testCommand("faucet", liquidAddress, liquid)
	if err == nil {
		t.Fatal("Should return error when trying logging liquid services if not running")
	}

	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	if err := testCommand("stop", "", delete); err != nil {
		t.Fatal(err)
	}
}
