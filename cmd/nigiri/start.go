package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/docker"
	"github.com/vulpemventures/nigiri/internal/proxy"
)

// proxyServers holds references to running proxy servers so they can be shut down
var proxyServers []*proxy.Server

var ciFlag = cli.BoolFlag{
	Name:  "ci",
	Usage: "runs in headless mode without esplora for continuous integration environments",
	Value: false,
}

var rememberFlag = cli.BoolFlag{
	Name:  "remember",
	Usage: "remember the flags used in this command for future runs",
	Value: false,
}

var start = cli.Command{
	Name:   "start",
	Usage:  "start nigiri",
	Action: startAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&lnFlag,
		&arkFlag,
		&ciFlag,
		&rememberFlag,
	},
}

const savedFlagsFileName = "flags.json"

type savedFlags struct {
	Liquid bool `json:"liquid"`
	Ln     bool `json:"ln"`
	Ark    bool `json:"ark"`
	Ci     bool `json:"ci"`
}

func loadFlags(datadir string) (*savedFlags, error) {
	flagsFilePath := filepath.Join(datadir, savedFlagsFileName)
	data, err := os.ReadFile(flagsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &savedFlags{}, nil
		}
		return nil, fmt.Errorf("failed to read saved flags file: %w", err)
	}

	var flags savedFlags
	if err := json.Unmarshal(data, &flags); err != nil {
		return nil, fmt.Errorf("failed to parse saved flags file: %w", err)
	}
	return &flags, nil
}

func saveFlags(datadir string, flags *savedFlags) error {
	flagsFilePath := filepath.Join(datadir, savedFlagsFileName)
	data, err := json.MarshalIndent(flags, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal flags: %w", err)
	}

	if err := os.WriteFile(flagsFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write saved flags file: %w", err)
	}
	fmt.Printf("Flags remembered in %s\n", flagsFilePath)
	return nil
}

func startAction(ctx *cli.Context) error {
	if isRunning, _ := nigiriState.GetBool("running"); isRunning {
		return errors.New("nigiri is already running, please stop it first")
	}

	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	loadedFlags, err := loadFlags(datadir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load saved flags: %v\n", err)
		loadedFlags = &savedFlags{}
	}

	effectiveFlags := savedFlags{
		Liquid: loadedFlags.Liquid,
		Ln:     loadedFlags.Ln,
		Ark:    loadedFlags.Ark,
		Ci:     loadedFlags.Ci,
	}

	if ctx.IsSet("liquid") {
		effectiveFlags.Liquid = ctx.Bool("liquid")
	}
	if ctx.IsSet("ln") {
		effectiveFlags.Ln = ctx.Bool("ln")
	}
	if ctx.IsSet("ark") {
		effectiveFlags.Ark = ctx.Bool("ark")
	}
	if ctx.IsSet("ci") {
		effectiveFlags.Ci = ctx.Bool("ci")
	}

	if ctx.Bool("remember") {
		if err := saveFlags(datadir, &effectiveFlags); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save flags: %v\n", err)
		}
	}

	var services []string

	if effectiveFlags.Ci {
		services = []string{"bitcoin", "electrs"}

		if effectiveFlags.Liquid {
			services = append(services, "liquid", "electrs-liquid")
		}
	} else {
		services = []string{"bitcoin", "electrs", "esplora"}

		if effectiveFlags.Liquid {
			services = append(services, "liquid", "electrs-liquid", "esplora-liquid")
		}
	}

	if effectiveFlags.Ln {
		services = append(services, "lnd", "tap", "cln")
	}

	if effectiveFlags.Ark {
		services = append(services, "ark", "ark-wallet")
	}

	bashCmd := runDockerCompose(composePath, append([]string{"up", "-d"}, services...)...)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	// Start the embedded proxy server(s) to replace chopsticks containers
	registryPath := filepath.Join(datadir, "registry")
	if err := os.MkdirAll(registryPath, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Bitcoin proxy on :3000
	btcProxyCfg := proxy.NewConfig(
		proxy.WithListenAddr("0.0.0.0:3000"),
		proxy.WithElectrsAddr("localhost:30000"),
		proxy.WithRPCAddr("localhost", "18443"),
		proxy.WithRPCCredentials("admin1", "123"),
		proxy.WithChain("bitcoin"),
		proxy.WithFaucet(true),
		proxy.WithMining(true),
		proxy.WithLogger(false),
		proxy.WithRegistryPath(registryPath),
	)
	btcProxy := proxy.NewServer(btcProxyCfg)
	btcErrChan := btcProxy.StartAsync()
	proxyServers = append(proxyServers, btcProxy)

	// Check for immediate startup errors
	select {
	case err := <-btcErrChan:
		if err != nil {
			return fmt.Errorf("failed to start bitcoin proxy: %w", err)
		}
	case <-time.After(500 * time.Millisecond):
		// Server started successfully
	}
	log.Printf("🔌 Embedded proxy (bitcoin) listening on :3000")

	if effectiveFlags.Liquid {
		liqProxyCfg := proxy.NewConfig(
			proxy.WithListenAddr("0.0.0.0:3001"),
			proxy.WithElectrsAddr("localhost:30001"),
			proxy.WithRPCAddr("localhost", "18884"),
			proxy.WithRPCCredentials("admin1", "123"),
			proxy.WithChain("liquid"),
			proxy.WithFaucet(true),
			proxy.WithMining(true),
			proxy.WithLogger(false),
			proxy.WithRegistryPath(registryPath),
		)
		liqProxy := proxy.NewServer(liqProxyCfg)
		liqErrChan := liqProxy.StartAsync()
		proxyServers = append(proxyServers, liqProxy)

		select {
		case err := <-liqErrChan:
			if err != nil {
				return fmt.Errorf("failed to start liquid proxy: %w", err)
			}
		case <-time.After(500 * time.Millisecond):
		}
		log.Printf("🔌 Embedded proxy (liquid) listening on :3001")
	}

	fmt.Printf("🍣 nigiri configuration located at %s\n", nigiriState.FilePath())
	if err := nigiriState.Set(map[string]string{
		"running": strconv.FormatBool(true),
		"liquid":  strconv.FormatBool(effectiveFlags.Liquid),
		"ln":      strconv.FormatBool(effectiveFlags.Ln),
		"ark":     strconv.FormatBool(effectiveFlags.Ark),
		"ci":      strconv.FormatBool(effectiveFlags.Ci),
	}); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	client := docker.NewDefaultClient()
	endpoints, err := client.GetEndpoints(composePath)
	if err != nil {
		return fmt.Errorf("failed to get endpoints: %w", err)
	}

	// Filter endpoints based on *effective* flags
	filteredEndpoints := make(map[string]string)
	for name, endpoint := range endpoints {
		if !effectiveFlags.Liquid && strings.Contains(name, "liquid") {
			continue
		}
		if !effectiveFlags.Ln && (strings.Contains(name, "lnd") || strings.Contains(name, "cln") || strings.Contains(name, "tap")) {
			continue
		}
		if !effectiveFlags.Ark && strings.Contains(name, "ark") {
			continue
		}

		filteredEndpoints[name] = endpoint
	}

	// Add proxy endpoints (replacing chopsticks)
	filteredEndpoints["chopsticks (embedded)"] = "localhost:3000"
	if effectiveFlags.Liquid {
		filteredEndpoints["chopsticks-liquid (embedded)"] = "localhost:3001"
	}

	// Display endpoints
	fmt.Println("\n🍜 ENDPOINTS")
	for name, endpoint := range filteredEndpoints {
		fmt.Printf("%s %s: %s\n",
			aurora.Green("✓"),
			aurora.Blue(name),
			endpoint,
		)
	}

	if effectiveFlags.Ark {
		// Wait for nbxplorer to sync
		if err := waitForNbxplorerSync(datadir); err != nil {
			return err
		}

		// Wait for ark containers to start
		if err := waitForArkContainers(client); err != nil {
			return err
		}

		// Setup arkd
		done := make(chan bool)
		go spinner(done, "setting up arkd...")

		if err := setupArk(datadir); err != nil {
			done <- true
			return fmt.Errorf("failed to setup Ark: %w", err)
		}

		done <- true
		fmt.Println("✓ arkd setup completed successfully!")
	}

	return nil
}

func spinner(done chan bool, message string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r%s\r", strings.Repeat(" ", len(message)+3))
			return
		default:
			fmt.Printf("\r%s %s", frames[i], message)
			time.Sleep(100 * time.Millisecond)
			i = (i + 1) % len(frames)
		}
	}
}

func waitForNbxplorerSync(datadir string) error {
	done := make(chan bool)
	go spinner(done, "waiting for nbxplorer to sync...")

	signalFilePath := filepath.Join(datadir, "volumes", "nbxplorer", "btc_fully_synched")
	timeout := time.After(120 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			done <- true
			return fmt.Errorf("timeout waiting for nbxplorer to sync")
		case <-ticker.C:
			if _, err := os.Stat(signalFilePath); err == nil {
				done <- true
				fmt.Println("✓ nbxplorer synced successfully!")
				return nil
			}
		}
	}
}

func waitForArkContainers(client docker.Client) error {
	done := make(chan bool)
	go spinner(done, "waiting for ark containers to start...")

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			done <- true
			return fmt.Errorf("timeout waiting for ark containers to start")
		case <-ticker.C:
			walletRunning, _ := client.IsContainerRunning("ark-wallet")
			arkRunning, _ := client.IsContainerRunning("ark")
			if walletRunning && arkRunning {
				done <- true
				fmt.Println("✓ ark containers started successfully!")
				return nil
			}
		}
	}
}

func setupArk(datadir string) error {
	time.Sleep(8 * time.Second) // Wait for ark containers to start before trying to set up the wallet

	bashCmd := exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "create", "--password", "secret")
	output, err := bashCmd.CombinedOutput()
	if err != nil {
		// Check if wallet is already initialized; this is not an error
		if strings.Contains(string(output), "wallet already initialized") {
			fmt.Println("ℹ wallet already initialized, skipping creation")
		} else {
			return fmt.Errorf("failed to create wallet: %w\nOutput: %s", err, string(output))
		}
	}

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "unlock", "--password", "secret")
	output, err = bashCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %w\nOutput: %s", err, string(output))
	}
	time.Sleep(4 * time.Second)
	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "status")
	output, err = bashCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check wallet status: %w\nOutput: %s", err, string(output))
	}

	time.Sleep(10 * time.Second)

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "ark", "init", "--password", "secret", "--server-url", "localhost:7070", "--explorer", "http://host.docker.internal:3000")
	output, err = bashCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize ark client wallet: %w\nOutput: %s", err, string(output))
	}

	// faucet arkd wallet
	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "address")
	output, err = bashCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get wallet address: %w", err)
	}
	address := strings.TrimSpace(string(output))

	// Fund the address using nigiri faucet
	for i := 0; i < 10; i++ {
		bashCmd = exec.Command("nigiri", "--datadir", datadir, "faucet", address)
		if err := bashCmd.Run(); err != nil {
			return fmt.Errorf("failed to fund wallet address: %w", err)
		}
	}

	return nil
}
