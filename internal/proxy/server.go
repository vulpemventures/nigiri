package proxy

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Server wraps the HTTP server and router
type Server struct {
	httpServer *http.Server
	router     *Router
	config     Config
}

// NewServer creates a new proxy server
func NewServer(config Config) *Server {
	router := NewRouter(config)

	httpServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         config.ListenURL(),
		Handler:      router,
	}

	return &Server{
		httpServer: httpServer,
		router:     router,
		config:     config,
	}
}

// Start starts the proxy server (blocking)
func (s *Server) Start() error {
	// Initialize the router (create wallet, fund faucet, etc.)
	if err := s.router.Initialize(); err != nil {
		log.Printf("Warning: router initialization had issues: %v", err)
		// Don't fail - the node might not be ready yet
	}

	log.Printf("Starting proxy server on %s", s.config.ListenURL())
	log.Printf("Configuration: chain=%s, faucet=%v, mining=%v, logger=%v",
		s.config.Chain(),
		s.config.IsFaucetEnabled(),
		s.config.IsMiningEnabled(),
		s.config.IsLoggerEnabled(),
	)

	if !s.config.IsTLSEnabled() {
		return s.httpServer.ListenAndServe()
	}

	// TLS with auto-cert
	dataDir := "."
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(dataDir),
	}
	s.httpServer.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

	return s.httpServer.ListenAndServeTLS("", "")
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// StartAsync starts the proxy server in a goroutine
func (s *Server) StartAsync() chan error {
	errChan := make(chan error, 1)

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	return errChan
}
