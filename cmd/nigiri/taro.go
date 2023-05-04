package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var taro = cli.Command{
	Name:   "taro",
	Usage:  "invoke taro command line interface",
	Action: taroAction,
}

func taroAction(ctx *cli.Context) error {

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
	rpcArgs := []string{"exec", ttyOption, "taro", "tarocli", "--network=" + network}
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
