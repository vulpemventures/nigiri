package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/pkg/state"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	defaultDataDir = btcutil.AppDataDir("nigiri-new", false)

	statePath    = filepath.Join(defaultDataDir, "nigiri.config.json")
	initialState = map[string]string{
		"attachliquid": strconv.FormatBool(false),
		"datadir":      defaultDataDir,
		"network":      "regtest",
		"ready":        strconv.FormatBool(false),
		"running":      strconv.FormatBool(false),
	}

	nigiriState = state.New(statePath, initialState)

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
		&start,
		&stop,
		&rpc,
		&logs,
		&mint,
		&push,
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
		return filepath.Join(defaultDataDir, regtestLiquidCompose)
	}

	return filepath.Join(defaultDataDir, regtestCompose)
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
	if err := makeDirectoryIfNotExists(filepath.Join(defaultDataDir, "volumes", "bitcoin")); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(filepath.Join(defaultDataDir, "volumes", "elements")); err != nil {
		return err
	}

	// copy resources into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", regtestCompose),
		filepath.Join(defaultDataDir, regtestCompose),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", regtestLiquidCompose),
		filepath.Join(defaultDataDir, regtestLiquidCompose),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "bitcoin.conf"),
		filepath.Join(defaultDataDir, "volumes", "bitcoin", "bitcoin.conf"),
	); err != nil {
		return err
	}

	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "elements.conf"),
		filepath.Join(defaultDataDir, "volumes", "elements", "elements.conf"),
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
