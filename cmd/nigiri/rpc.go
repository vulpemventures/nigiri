package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var rpc = cli.Command{
	Name:   "rpc",
	Usage:  "invoke bitcoin-cli or elements-cli",
	Action: rpcAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&cli.StringFlag{
			Name:  "rpcwallet",
			Usage: "rpcwallet to be used for node JSONRPC commands",
			Value: "",
		},
	},
}

func rpcAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	isLiquid := ctx.Bool("liquid")
	rpcWallet := ctx.String("rpcwallet")

	rpcArgs := []string{"exec", "bitcoin", "bitcoin-cli", "-datadir=config", "-rpcwallet=" + rpcWallet}
	if isLiquid {
		rpcArgs = []string{"exec", "liquid", "elements-cli", "-datadir=config", "-rpcwallet=" + rpcWallet}
	}
	cmdArgs := append(rpcArgs, ctx.Args().Slice()...)
	bashCmd := exec.Command("docker", cmdArgs...)
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	if err := bashCmd.Run(); err != nil {
		return err
	}

	return nil
}
