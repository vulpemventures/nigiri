package proxy

import (
	"os"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}
}

func TestRegistry_AddAndGetEntry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	asset := "asset123"
	issuanceInput := map[string]interface{}{
		"txid": "tx123",
		"vin":  float64(0),
	}
	contract := map[string]interface{}{
		"name":      "TestAsset",
		"ticker":    "TEST",
		"precision": 8,
	}

	err = registry.AddEntry(asset, issuanceInput, contract)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get the entry back
	entry, err := registry.GetEntry(asset)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	if entry["name"] != "TestAsset" {
		t.Errorf("Expected name 'TestAsset', got '%v'", entry["name"])
	}

	if entry["ticker"] != "TEST" {
		t.Errorf("Expected ticker 'TEST', got '%v'", entry["ticker"])
	}
}

func TestRegistry_AddDuplicateEntry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	asset := "asset123"
	issuanceInput := map[string]interface{}{
		"txid": "tx123",
		"vin":  float64(0),
	}
	contract := map[string]interface{}{
		"name":      "TestAsset",
		"ticker":    "TEST",
		"precision": 8,
	}

	// Add first entry
	err = registry.AddEntry(asset, issuanceInput, contract)
	if err != nil {
		t.Fatalf("Failed to add first entry: %v", err)
	}

	// Try to add duplicate
	err = registry.AddEntry(asset, issuanceInput, contract)
	if err == nil {
		t.Error("Expected error when adding duplicate entry")
	}
}

func TestRegistry_GetNonExistentEntry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	entry, err := registry.GetEntry("nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(entry) != 0 {
		t.Errorf("Expected empty entry for nonexistent asset, got %v", entry)
	}
}

func TestRegistry_GetEntries_All(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add multiple entries - use simple asset IDs without slashes
	assets := []struct {
		id     string
		name   string
		ticker string
	}{
		{"assetA123", "Asset1", "A"},
		{"assetB456", "Asset2", "B"},
	}

	for _, a := range assets {
		issuanceInput := map[string]interface{}{
			"txid": "tx" + a.id,
			"vin":  float64(0),
		}
		contract := map[string]interface{}{
			"name":      a.name,
			"ticker":    a.ticker,
			"precision": 8,
		}
		err = registry.AddEntry(a.id, issuanceInput, contract)
		if err != nil {
			t.Fatalf("Failed to add entry %s: %v", a.id, err)
		}
	}

	// Get all entries (empty slice means all)
	entries, err := registry.GetEntries([]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestRegistry_GetEntries_Specific(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add entry
	asset := "specificasset"
	issuanceInput := map[string]interface{}{
		"txid": "tx123",
		"vin":  float64(0),
	}
	contract := map[string]interface{}{
		"name":      "SpecificAsset",
		"ticker":    "SPEC",
		"precision": 8,
	}
	err = registry.AddEntry(asset, issuanceInput, contract)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get specific entry
	entries, err := registry.GetEntries([]interface{}{asset})
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0]["name"] != "SpecificAsset" {
		t.Errorf("Expected name 'SpecificAsset', got '%v'", entries[0]["name"])
	}
}
