package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var ark = cli.Command{
	Name:   "ark",
	Usage:  "invoke ark client commands",
	Action: arkAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "print version",
		},
	},
}

var arkd = cli.Command{
	Name:   "arkd",
	Usage:  "invoke arkd client commands",
	Action: arkdAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "print version",
		},
	},
}

func arkAction(ctx *cli.Context) error {
	if ctx.Bool("version") {
		return runArkCommand(ctx, "ark", "--version")
	}
	return runArkCommand(ctx, "ark", ctx.Args().Slice()...)
}

func arkdAction(ctx *cli.Context) error {
	if ctx.Bool("version") {
		return runArkCommand(ctx, "arkd", "--version")
	}
	args := ctx.Args().Slice()
	// Add default flags
	args = append([]string{"--no-macaroon"}, args...)
	return runArkCommand(ctx, "arkd", args...)
}

func runArkCommand(ctx *cli.Context, binary string, args ...string) error {
	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	isArkEnabled, _ := nigiriState.GetBool("ark")
	if !isArkEnabled {
		return errors.New("ark is not enabled. Start nigiri with --ark flag")
	}

	// Build the docker exec command
	cmd := []string{"exec", "ark", binary}

	// Pass through all arguments
	cmd = append(cmd, args...)

	// Run the command
	dockerCmd := exec.Command("docker", cmd...)
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	if err := dockerCmd.Run(); err != nil {
		// The error is already printed to stderr
		return err
	}

	return nil
}
