package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var versionCmd = cli.Command{
	Name:   "version",
	Action: versionAction,
}

func versionAction(ctx *cli.Context) error {
	fmt.Println("nigiri CLI version")
	fmt.Println(formatVersion())
	return nil
}
