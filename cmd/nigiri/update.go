package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
)

const defaultUpdateScriptURL = "https://getnigiri.vulpem.com"

var update = cli.Command{
	Name:   "update",
	Usage:  "check for updates and pull new docker images",
	Action: updateAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "force update the nigiri binary to the latest version",
		},
	},
}

func updateAction(ctx *cli.Context) error {
	if ctx.Bool("force") {
		return updateBinary(ctx)
	}

	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	bashCmd := runDockerCompose(composePath, "pull")
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}

func updateBinary(ctx *cli.Context) error {
	// Get the update script URL from environment variable or use default
	updateURL := os.Getenv("NIGIRI_TEST_UPDATE_SCRIPT_URL")
	if updateURL == "" {
		updateURL = defaultUpdateScriptURL
	}

	fmt.Printf("Downloading update script from %s...\n", updateURL)

	// Download the update script
	resp, err := http.Get(updateURL)
	if err != nil {
		return fmt.Errorf("failed to download update script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update script: HTTP %d", resp.StatusCode)
	}

	// Create a temporary file for the update script
	tmpFile, err := os.CreateTemp("", "nigiri-update-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write the script to the temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write update script: %w", err)
	}
	tmpFile.Close()

	// Make the script executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	fmt.Println("Executing update script...")

	// Execute the update script, replacing the current process
	return syscall.Exec(tmpPath, []string{tmpPath}, os.Environ())
}
