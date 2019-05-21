package cmd

import (
	"testing"
)

var (
	serviceList = []string{"node", "electrs", "esplora", "chopsticks"}
)

func TestLogBitcoinServices(t *testing.T) {
	if err := testCommand("start", "", bitcoin); err != nil {
		t.Fatal(err)
	}

	for _, service := range serviceList {
		if err := testCommand("logs", service, bitcoin); err != nil {
			t.Fatal(err)
		}
	}

	if err := testCommand("stop", "", delete); err != nil {
		t.Fatal(err)
	}
}

func TestLogLiquidServices(t *testing.T) {
	if err := testCommand("start", "", liquid); err != nil {
		t.Fatal(err)
	}

	for _, service := range serviceList {
		if err := testCommand("logs", service, liquid); err != nil {
			t.Fatal(err)
		}
	}

	if err := testCommand("stop", "", delete); err != nil {
		t.Fatal(err)
	}
}

func TestLogShouldFail(t *testing.T) {
	expectedError := "Nigiri is not running"

	err := testCommand("logs", serviceList[0], bitcoin)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}

	err = testCommand("logs", serviceList[0], liquid)
	if err == nil {
		t.Fatal("Should return error when Nigiri is stopped")
	}
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err)
	}
}

func TestStartBitcoinAndLogNigiriServicesShouldFail(t *testing.T) {
	if err := testCommand("start", "", bitcoin); err != nil {
		t.Fatal(err)
	}

	expectedError := "Nigiri has been started with no Liquid sidechain.\nPlease stop and restart it using the --liquid flag"

	err := testCommand("logs", serviceList[0], liquid)
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
