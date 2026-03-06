package proxy

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	cfg := NewConfig(WithListenAddr("localhost:0"))
	server := NewServer(cfg)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config == nil {
		t.Error("Expected config to be set")
	}

	if server.router == nil {
		t.Error("Expected router to be set")
	}

	if server.httpServer == nil {
		t.Error("Expected httpServer to be set")
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	// Use a high port number to avoid conflicts
	cfg := NewConfig(
		WithListenAddr("127.0.0.1:19999"),
		WithFaucet(false), // Disable faucet to avoid RPC connection issues
	)
	server := NewServer(cfg)

	// Start server async
	errChan := server.StartAsync()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is responding
	resp, err := http.Get("http://127.0.0.1:19999/getnewaddress")
	if err == nil {
		resp.Body.Close()
		// We expect an error since there's no real RPC server, but the point is the HTTP server is up
	}

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown error: %v", err)
	}

	// Check that no error was returned from start
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Server returned unexpected error: %v", err)
		}
	default:
		// Channel might be closed or empty, that's fine
	}
}

func TestServer_StartAsync(t *testing.T) {
	cfg := NewConfig(
		WithListenAddr("127.0.0.1:19998"),
		WithFaucet(false),
	)
	server := NewServer(cfg)

	errChan := server.StartAsync()

	// Verify errChan is non-nil
	if errChan == nil {
		t.Fatal("Expected non-nil error channel")
	}

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
