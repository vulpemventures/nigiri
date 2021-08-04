package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/pkg/state"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	nigiriDataDir = config.DefaultDatadir
	nigiriState   = state.New(config.DefaultPath, config.InitialState)

	regtestCompose       = "docker-compose-regtest.yml"
	regtestLiquidCompose = "docker-compose-regtest-liquid.yml"
)

var liquidFlag = cli.BoolFlag{
	Name:  "liquid",
	Usage: "enable liquid",
	Value: false,
}

//go:embed resources/docker-compose-regtest.yml
//go:embed resources/docker-compose-regtest-liquid.yml
//go:embed resources/bitcoin.conf
//go:embed resources/elements.conf
var f embed.FS

func init() {
	dataDir := cleanAndExpandPath(os.Getenv("NIGIRI_DATADIR"))
	if len(dataDir) > 0 {
		nigiriState = state.New(filepath.Join(dataDir, config.DefaultName), config.InitialState)
		nigiriDataDir = dataDir
	}

	if err := provisionResourcesToDatadir(); err != nil {
		fatal(err)
	}
}

func main() {
	app := cli.NewApp()

	app.Version = formatVersion()
	app.Name = "nigiri CLI"
	app.Usage = "create your dockerized environment with a bitcoin and liquid node, with a block explorer and developer tools"
	app.Commands = append(
		app.Commands,
		&rpc,
		&stop,
		&logs,
		&mint,
		&push,
		&start,
		&update,
		&faucet,
	)

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "[nigiri] %v\n", err)
	os.Exit(1)
}

func getCompose(isLiquid bool) string {
	if isLiquid {
		return filepath.Join(nigiriDataDir, regtestLiquidCompose)
	}

	return filepath.Join(nigiriDataDir, regtestCompose)
}

// Provisioning Nigiri reosurces
func provisionResourcesToDatadir() error {

	isReady, err := nigiriState.GetBool("ready")
	if err != nil {
		return err
	}

	if isReady {
		return nil
	}

	// create folders in volumes/{bitcoin,elements} for node datadirs
	if err := makeDirectoryIfNotExists(filepath.Join(nigiriDataDir, "volumes", "bitcoin")); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(filepath.Join(nigiriDataDir, "volumes", "elements")); err != nil {
		return err
	}

	// copy resources into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", regtestCompose),
		filepath.Join(nigiriDataDir, regtestCompose),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", regtestLiquidCompose),
		filepath.Join(nigiriDataDir, regtestLiquidCompose),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "bitcoin.conf"),
		filepath.Join(nigiriDataDir, "volumes", "bitcoin", "bitcoin.conf"),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "elements.conf"),
		filepath.Join(nigiriDataDir, "volumes", "elements", "elements.conf"),
	); err != nil {
		return err
	}

	if err := nigiriState.Set(map[string]string{"ready": strconv.FormatBool(true)}); err != nil {
		return err
	}

	return nil
}

func formatVersion() string {
	return fmt.Sprintf(
		"\nVersion: %s\nCommit: %s\nDate: %s",
		version, commit, date,
	)
}

func copyFromResourcesToDatadir(src string, dest string) error {
	data, err := f.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read embed: %w", err)
	}
	err = ioutil.WriteFile(dest, data, 0777)
	if err != nil {
		return fmt.Errorf("write %s to %s: %w", src, dest, err)
	}

	return nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModeDir|0755)
	}
	return nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func cleanAndExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		u, err := user.Current()
		if err == nil {
			homeDir = u.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}

		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}
