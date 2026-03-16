package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/proxy"
)

var addrFlag = cli.StringFlag{
	Name:  "addr",
	Usage: "listen address for the proxy server",
	Value: "localhost:3000",
}

var electrsAddrFlag = cli.StringFlag{
	Name:  "electrs-addr",
	Usage: "electrs HTTP server address",
	Value: "localhost:30000",
}

var rpcAddrFlag = cli.StringFlag{
	Name:  "rpc-addr",
	Usage: "RPC server address (host:port)",
	Value: "localhost:18443",
}

var rpcCookieFlag = cli.StringFlag{
	Name:  "rpc-cookie",
	Usage: "RPC server user:password",
	Value: "admin1:123",
}

var chainFlag = cli.StringFlag{
	Name:  "chain",
	Usage: "chain type (bitcoin or liquid)",
	Value: "bitcoin",
}

var useFaucetFlag = cli.BoolFlag{
	Name:  "use-faucet",
	Usage: "enable faucet endpoint",
	Value: true,
}

var useMiningFlag = cli.BoolFlag{
	Name:  "use-mining",
	Usage: "mine blocks after broadcasts",
	Value: true,
}

var useLoggerFlag = cli.BoolFlag{
	Name:  "use-logger",
	Usage: "log all requests",
	Value: false,
}

var serve = cli.Command{
	Name:  "serve",
	Usage: "start the chopsticks proxy server",
	Description: `Start an embedded HTTP proxy server that provides:
   - /faucet endpoint for funding addresses
   - /mint endpoint for issuing Liquid assets
   - /registry endpoint for asset metadata
   - Auto-mining after transaction broadcasts
   - Proxy to electrs for all other requests

   This replaces the need for the separate nigiri-chopsticks Docker container.`,
	Action: serveAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&addrFlag,
		&electrsAddrFlag,
		&rpcAddrFlag,
		&rpcCookieFlag,
		&chainFlag,
		&useFaucetFlag,
		&useMiningFlag,
		&useLoggerFlag,
	},
}

func serveAction(ctx *cli.Context) error {
	datadir := ctx.String("datadir")

	// Parse RPC address
	rpcAddr := ctx.String("rpc-addr")
	rpcHost, rpcPort := "localhost", "18443"
	if idx := lastIndex(rpcAddr, ':'); idx != -1 {
		rpcHost = rpcAddr[:idx]
		rpcPort = rpcAddr[idx+1:]
	}

	// Parse RPC cookie
	rpcCookie := ctx.String("rpc-cookie")
	rpcUser, rpcPass := "admin1", "123"
	if idx := lastIndex(rpcCookie, ':'); idx != -1 {
		rpcUser = rpcCookie[:idx]
		rpcPass = rpcCookie[idx+1:]
	}

	// Determine chain and adjust defaults for liquid
	chain := ctx.String("chain")
	if ctx.Bool("liquid") {
		chain = "liquid"
	}

	listenAddr := ctx.String("addr")
	electrsAddr := ctx.String("electrs-addr")

	// Adjust defaults for liquid if --liquid flag is set
	if chain == "liquid" && !ctx.IsSet("addr") {
		listenAddr = "localhost:3001"
	}
	if chain == "liquid" && !ctx.IsSet("electrs-addr") {
		electrsAddr = "localhost:30001"
	}
	if chain == "liquid" && !ctx.IsSet("rpc-addr") {
		rpcHost = "localhost"
		rpcPort = "18884"
	}

	// Registry path for Liquid assets
	registryPath := filepath.Join(datadir, "registry")
	if err := os.MkdirAll(registryPath, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Build config
	cfg := proxy.NewConfig(
		proxy.WithListenAddr(listenAddr),
		proxy.WithElectrsAddr(electrsAddr),
		proxy.WithRPCAddr(rpcHost, rpcPort),
		proxy.WithRPCCredentials(rpcUser, rpcPass),
		proxy.WithChain(chain),
		proxy.WithFaucet(ctx.Bool("use-faucet")),
		proxy.WithMining(ctx.Bool("use-mining")),
		proxy.WithLogger(ctx.Bool("use-logger")),
		proxy.WithRegistryPath(registryPath),
	)

	// Create and start server
	server := proxy.NewServer(cfg)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := server.StartAsync()

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigChan:
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		if err := server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
	}

	return nil
}

// lastIndex returns the last index of sep in s, or -1 if not found
func lastIndex(s string, sep byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}

var serveDefaultsFile = filepath.Join(config.DefaultDatadir, "serve-defaults.json")
