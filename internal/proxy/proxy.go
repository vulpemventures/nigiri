package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	BitcoinRPCPort  = "28443"
	ElectrumRPCPort = "50001"
	EsploraAPIPort  = "30001"

	ProxyBitcoinPort  = "18443"
	ProxyElectrumPort = "50000"
	ProxyEsploraPort  = "30000"
)

func Start() error {
	// Bitcoin Core JSON-RPC proxy
	go func() {
		if err := startTCPProxy(BitcoinRPCPort, ProxyBitcoinPort, "Bitcoin Core"); err != nil {
			fmt.Printf("Error starting Bitcoin Core proxy: %v\n", err)
		}
	}()

	// ElectrumX JSON-RPC proxy
	go func() {
		if err := startTCPProxy(ElectrumRPCPort, ProxyElectrumPort, "ElectrumX"); err != nil {
			fmt.Printf("Error starting ElectrumX proxy: %v\n", err)
		}
	}()

	// Esplora API proxy (keeping HTTP for now)
	return startEsploraProxy(EsploraAPIPort, ProxyEsploraPort)
}

func startTCPProxy(targetPort, listenPort, serviceName string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", listenPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	log.Printf("Started %s proxy on port %s", serviceName, listenPort)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(clientConn, targetPort, serviceName)
	}
}

func handleConnection(clientConn net.Conn, targetPort, serviceName string) {
	defer clientConn.Close()

	targetConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", targetPort))
	if err != nil {
		log.Printf("Error connecting to target service: %v", err)
		return
	}
	defer targetConn.Close()

	// Client to target
	go func() {
		if err := proxyData(clientConn, targetConn, serviceName, "request"); err != nil {
			log.Printf("Error proxying data from client to target: %v", err)
		}
	}()

	// Target to client
	if err := proxyData(targetConn, clientConn, serviceName, "response"); err != nil {
		log.Printf("Error proxying data from target to client: %v", err)
	}
}

func proxyData(src, dst net.Conn, serviceName, direction string) error {
	reader := bufio.NewReader(src)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Try to parse as JSON-RPC
		var jsonRPC map[string]interface{}
		if err := json.Unmarshal([]byte(message), &jsonRPC); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonRPC, "", "  ")
			log.Printf("%s %s (JSON-RPC):\n%s", serviceName, direction, string(prettyJSON))
		} else {
			// If not JSON-RPC, just print the raw message
			log.Printf("%s %s (raw):\n%s", serviceName, direction, strings.TrimSpace(message))
		}

		// Forward the message to the destination
		if _, err := dst.Write([]byte(message)); err != nil {
			return err
		}
	}
}

func startEsploraProxy(targetPort, listenPort string) error {
	target, err := url.Parse(fmt.Sprintf("http://localhost:%s", targetPort))
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/tx") {
			log.Println("Intercepted POST /tx request")
			// You can add more logging or processing here
		}
	}

	return http.ListenAndServe(fmt.Sprintf(":%s", listenPort), nil)
}
