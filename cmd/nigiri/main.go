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
	"github.com/vulpemventures/nigiri/internal/state"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	nigiriState = state.New(config.DefaultPath, config.InitialState)
)

var liquidFlag = cli.BoolFlag{
	Name:  "liquid",
	Usage: "enable liquid",
	Value: false,
}

var lnFlag = cli.BoolFlag{
	Name:  "ln",
	Usage: "enable Lightning Network",
	Value: false,
}

var datadirFlag = cli.StringFlag{
	Name:  "datadir",
	Usage: "use different data directory",
	Value: config.DefaultDatadir,
}

//go:embed resources/docker-compose.yml
//go:embed resources/bitcoin.conf
//go:embed resources/elements.conf
//go:embed resources/lnd.conf
//go:embed resources/lightning.conf
var f embed.FS

func main() {
	app := cli.NewApp()

	app.Version = formatVersion()
	app.Name = "nigiri CLI"
	app.Usage = "create your dockerized environment with a bitcoin and liquid node, with a block explorer and developer tools"
	app.Flags = append(app.Flags, &datadirFlag)
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
		&versionCmd,
	)

	app.Before = func(ctx *cli.Context) error {

		dataDir := config.DefaultDatadir

		if ctx.IsSet("datadir") {
			dataDir = cleanAndExpandPath(ctx.String("datadir"))
			nigiriState = state.New(filepath.Join(dataDir, config.DefaultName), config.InitialState)
		}

		if err := provisionResourcesToDatadir(dataDir); err != nil {
			return err
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "[nigiri] %v\n", err)
	os.Exit(1)
}

// Provisioning Nigiri reosurces
func provisionResourcesToDatadir(datadir string) error {

	isReady, err := nigiriState.GetBool("ready")
	if err != nil {
		return err
	}

	if isReady {
		return nil
	}

	// create folders in volumes/{bitcoin,elements} for node datadirs
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, "volumes", "bitcoin")); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, "volumes", "elements")); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, "volumes", "lnd")); err != nil {
		return err
	}
	if err := makeDirectoryIfNotExists(filepath.Join(datadir, "volumes", "lnd2")); err != nil {
		return err
	}

	// copy docker compose into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", config.DefaultCompose),
		filepath.Join(datadir, config.DefaultCompose),
	); err != nil {
		return err
	}

	// copy bitcoin.conf into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "bitcoin.conf"),
		filepath.Join(datadir, "volumes", "bitcoin", "bitcoin.conf"),
	); err != nil {
		return err
	}

	// copy elements.conf into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "elements.conf"),
		filepath.Join(datadir, "volumes", "elements", "elements.conf"),
	); err != nil {
		return err
	}

	// copy lnd.conf into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "lnd.conf"),
		filepath.Join(datadir, "volumes", "lnd", "lnd.conf"),
	); err != nil {
		return err
	}

	// copy second lnd.conf into the Nigiri data directory
	if err := copyFromResourcesToDatadir(
		filepath.Join("resources", "lnd.conf"),
		filepath.Join(datadir, "volumes", "lnd2", "lnd.conf"),
	); err != nil {
		return err
	}

	/* 	// copy lightning.conf into the Nigiri data directory
	   	if err := copyFromResourcesToDatadir(
	   		filepath.Join("resources", "lightning.conf"),
	   		filepath.Join(datadir, "volumes", "lightningd", "lightning.conf"),
	   	); err != nil {
	   		return err
	   	} */

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
