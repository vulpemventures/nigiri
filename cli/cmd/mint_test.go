package cmd

import (
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

func TestMintCommand(t *testing.T) {
	testStart(t, liquid)

	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 1000", liquid); err != nil {
		t.Fatal(err)
	}
	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 2000 Test", liquid); err != nil {
		t.Fatal(err)
	}
	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 3000 L-BTC LBTC", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestMintShouldFail(t *testing.T) {
	expectedError := constants.ErrNigiriNotRunning.Error()

	err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 1000", liquid)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func TestStartBitcoinAndMintShouldFail(t *testing.T) {
	testStart(t, bitcoin)

	expectedError := constants.ErrNigiriLiquidNotEnabled.Error()

	err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 1000", liquid)
	if err == nil {
		t.Fatal("Should return error when trying logging liquid services if not running")
	}

	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	testDelete(t)
}
