package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var tap = cli.Command{
	Name:   "tap",
	Usage:  "invoke tap command line interface",
	Action: tapAction,
}

func tapAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	network, err := nigiriState.GetString("network")
	if err != nil {
		return err
	}

	isCi, err := nigiriState.GetBool("ci")
	if err != nil {
		return err
	}

	ttyOption := "-it"
	if isCi {
		ttyOption = "-i"
	}
	rpcArgs := []string{"exec", ttyOption, "tap", "tapcli", "--network=" + network, "--tapddir=/data/.tapd"}
	cmdArgs := append(rpcArgs, ctx.Args().Slice()...)
	bashCmd := exec.Command("docker", cmdArgs...)
	bashCmd.Stdin = os.Stdin
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
