package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	cfg := NewTestConfig()
	router := NewRouter(cfg)

	if router == nil {
		t.Fatal("Expected non-nil Router")
	}

	if router.Config == nil {
		t.Error("Expected Config to be set")
	}

	if router.RPCClient == nil {
		t.Error("Expected RPCClient to be set")
	}
}

func TestHandleAddressRequest_MockRPC(t *testing.T) {
	// Create a mock RPC server
	mockRPC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to check the method
		body, _ := ioutil.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		if req["method"] == "getnewaddress" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "bcrt1qtest123",
				"error":  nil,
				"id":     req["id"],
			})
			return
		}

		http.Error(w, "unknown method", http.StatusBadRequest)
	}))
	defer mockRPC.Close()

	// Extract host and port from mock server URL
	cfg := NewConfig(
		WithListenAddr("localhost:0"),
		WithElectrsAddr("localhost:30000"),
		WithRPCAddr(mockRPC.URL[7:], ""), // Remove "http://" prefix
		WithFaucet(false),                 // Disable faucet to avoid initialization issues
	)

	// Create a custom router with mock RPC client
	router := NewRouter(cfg)
	// Replace the RPC client with one pointing to our mock
	router.RPCClient, _ = NewRPCClient(mockRPC.URL+"/wallet/", false, 10)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/getnewaddress", nil)
	w := httptest.NewRecorder()

	// Call the handler
	router.HandleAddressRequest(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["address"] != "bcrt1qtest123" {
		t.Errorf("Expected address bcrt1qtest123, got %s", result["address"])
	}
}

func TestHandleFaucetRequest_MissingAddress(t *testing.T) {
	cfg := NewConfig(WithFaucet(true))
	router := NewRouter(cfg)
	router.Faucet = NewFaucet("http://localhost:18443", router.RPCClient)

	// Create request without address
	body := bytes.NewBufferString(`{"amount": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/faucet", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleFaucetRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(respBody, []byte("missing address")) {
		t.Errorf("Expected error about missing address, got: %s", string(respBody))
	}
}

func TestHandleMintRequest_MissingAddress(t *testing.T) {
	cfg := NewConfig(
		WithChain("liquid"),
		WithFaucet(true),
	)
	router := NewRouter(cfg)
	router.Faucet = NewFaucet("http://localhost:18884", router.RPCClient)

	// Create request without address
	body := bytes.NewBufferString(`{"amount": 1}`)
	req := httptest.NewRequest(http.MethodPost, "/mint", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleMintRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleMintRequest_MissingAmount(t *testing.T) {
	cfg := NewConfig(
		WithChain("liquid"),
		WithFaucet(true),
	)
	router := NewRouter(cfg)
	router.Faucet = NewFaucet("http://localhost:18884", router.RPCClient)

	// Create request without amount
	body := bytes.NewBufferString(`{"address": "el1qqtest123"}`)
	req := httptest.NewRequest(http.MethodPost, "/mint", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleMintRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(respBody, []byte("missing amount")) {
		t.Errorf("Expected error about missing amount, got: %s", string(respBody))
	}
}

func TestHandleRegistryRequest_MissingAssets(t *testing.T) {
	cfg := NewConfig(
		WithChain("liquid"),
		WithFaucet(true),
	)
	router := NewRouter(cfg)

	// Create request without assets
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/registry", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleRegistryRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleSubmitPackageRequest_InvalidJSON(t *testing.T) {
	cfg := NewTestConfig()
	router := NewRouter(cfg)

	// Create request with invalid JSON
	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/txs/package", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.HandleSubmitPackageRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestOptionsRequest(t *testing.T) {
	cfg := NewTestConfig()
	router := NewRouter(cfg)

	req := httptest.NewRequest(http.MethodOptions, "/faucet", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", resp.StatusCode)
	}

	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected CORS header *, got %s", cors)
	}
}

func TestParseRequestBody(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString(`{"address": "test", "amount": 1.5}`))
	result := parseRequestBody(body)

	if result["address"] != "test" {
		t.Errorf("Expected address 'test', got %v", result["address"])
	}

	if result["amount"] != 1.5 {
		t.Errorf("Expected amount 1.5, got %v", result["amount"])
	}
}

func TestParseRequestBody_Empty(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString(``))
	result := parseRequestBody(body)

	// Empty body results in nil map from json decode, which is expected behavior
	if result != nil && len(result) > 0 {
		t.Error("Expected empty or nil map for empty body")
	}
}
