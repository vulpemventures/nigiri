package cmd

import (
	"testing"

	"github.com/vulpemventures/nigiri/cli/constants"
)

func TestMintOneArg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 1000", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestMintTwoArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 2000 Test", liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestMintThreeArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)

	if err := testCommand("mint", "ert1q90dz89u8eudeswzynl3p2jke564ejc2cnfcwuq 3000 TEST TST", liquid); err != nil {
		t.Fatal(err)
	}
	testDelete(t)

}

func TestStartBitcoinAndMintShouldFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
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
