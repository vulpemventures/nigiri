package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
)

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

var composeFlag = cli.StringFlag{
	Name:  "compose",
	Usage: "use a custom docker-compose file instead of the embedded one",
	Value: "",
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
		&composeFlag,
	},
}

const savedFlagsFileName = "flags.json"

type savedFlags struct {
	Liquid  bool   `json:"liquid"`
	Ln      bool   `json:"ln"`
	Ark     bool   `json:"ark"`
	Ci      bool   `json:"ci"`
	Compose string `json:"compose,omitempty"`
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
	
	loadedFlags, err := loadFlags(datadir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load saved flags: %v\n", err)
		loadedFlags = &savedFlags{}
	}

	effectiveFlags := savedFlags{
		Liquid:  loadedFlags.Liquid,
		Ln:      loadedFlags.Ln,
		Ark:     loadedFlags.Ark,
		Ci:      loadedFlags.Ci,
		Compose: loadedFlags.Compose,
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
	if ctx.IsSet("compose") {
		effectiveFlags.Compose = ctx.String("compose")
	}

	// Use custom compose file if provided (either from CLI or saved flags), otherwise use default
	composePath := effectiveFlags.Compose
	if composePath == "" {
		composePath = filepath.Join(datadir, config.DefaultCompose)
	} else {
		// Expand and clean the custom compose path
		composePath = cleanAndExpandPath(composePath)
		
		// Verify the custom compose file exists
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			return fmt.Errorf("custom compose file not found: %s", composePath)
		}
	}

	if ctx.Bool("remember") {
		if err := saveFlags(datadir, &effectiveFlags); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save flags: %v\n", err)
		}
	}

	var services []string

	if effectiveFlags.Ci {
		services = []string{"bitcoin", "electrs", "chopsticks"}

		if effectiveFlags.Liquid {
			services = append(services, "liquid", "electrs-liquid", "chopsticks-liquid")
		}
	} else {
		services = []string{"bitcoin", "electrs", "chopsticks", "esplora"}

		if effectiveFlags.Liquid {
			services = append(services, "liquid", "electrs-liquid", "chopsticks-liquid", "esplora-liquid")
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
		done := make(chan bool)
		go spinner(done, "setting up arkd...")

		time.Sleep(4 * time.Second) // give time for the container to start
		if err := setupArk(); err != nil {
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

func setupArk() error {
	bashCmd := exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "create", "--password", "secret")
	bashCmd.Run()

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "unlock", "--password", "secret")
	if err := bashCmd.Run(); err != nil {
		return fmt.Errorf("failed to unlock wallet: %w", err)
	}

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "status")
	if err := bashCmd.Run(); err != nil {
		return fmt.Errorf("failed to check wallet status: %w", err)
	}

	time.Sleep(10 * time.Second)

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "ark", "init", "--network", "regtest", "--password", "secret", "--server-url", "localhost:7070", "--explorer", "http://chopsticks:3000")
	bashCmd.Run() // Ignore error as wallet might already exist

	// faucet arkd wallet
	bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "address")
	output, err := bashCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get wallet address: %w", err)
	}
	address := strings.TrimSpace(string(output))

	// Fund the address using nigiri faucet
	for i := 0; i < 10; i++ {
		bashCmd = exec.Command("nigiri", "faucet", address)
		if err := bashCmd.Run(); err != nil {
			return fmt.Errorf("failed to fund wallet address: %w", err)
		}
	}

	return nil
}
