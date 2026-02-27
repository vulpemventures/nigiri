package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRPCClient(t *testing.T) {
	client, err := NewRPCClient("http://localhost:18443", false, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestNewRPCClient_SSL(t *testing.T) {
	client, err := NewRPCClient("https://localhost:18443", true, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestRPCClient_Call(t *testing.T) {
	// Create a mock RPC server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "test_result",
			"error":  nil,
			"id":     12345,
		})
	}))
	defer mockServer.Close()

	client, _ := NewRPCClient(mockServer.URL, false, 10)
	status, resp, err := client.Call("getblockcount", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	var result string
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result != "test_result" {
		t.Errorf("Expected result 'test_result', got '%s'", result)
	}
}

func TestRPCClient_Call_Error(t *testing.T) {
	// Create a mock RPC server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -1,
				"message": "test error",
			},
		})
	}))
	defer mockServer.Close()

	client, _ := NewRPCClient(mockServer.URL, false, 10)
	_, _, err := client.Call("getblockcount", nil)

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestHandleRPCRequest(t *testing.T) {
	// Create a mock RPC server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": float64(100),
			"error":  nil,
			"id":     12345,
		})
	}))
	defer mockServer.Close()

	client, _ := NewRPCClient(mockServer.URL, false, 10)
	status, result, err := HandleRPCRequest(client, "getblockcount", nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}

	if result.(float64) != 100 {
		t.Errorf("Expected result 100, got %v", result)
	}
}

func TestHandleRPCRequest_WithParams(t *testing.T) {
	var receivedParams []interface{}

	// Create a mock RPC server that captures params
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		receivedParams = req["params"].([]interface{})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "tx123",
			"error":  nil,
			"id":     12345,
		})
	}))
	defer mockServer.Close()

	client, _ := NewRPCClient(mockServer.URL, false, 10)
	params := []interface{}{"bcrt1qtest", float64(1.5)}
	_, _, err := HandleRPCRequest(client, "sendtoaddress", params)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(receivedParams) != 2 {
		t.Errorf("Expected 2 params, got %d", len(receivedParams))
	}

	if receivedParams[0] != "bcrt1qtest" {
		t.Errorf("Expected first param 'bcrt1qtest', got %v", receivedParams[0])
	}
}
