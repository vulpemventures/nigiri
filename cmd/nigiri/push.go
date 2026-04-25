package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/urfave/cli/v2"
)

var push = cli.Command{
	Name:      "push",
	Usage:     "broadcast raw transaction",
	ArgsUsage: "<hex>",
	Action:    pushAction,
	Flags: []cli.Flag{
		&liquidFlag,
	},
}

func pushAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() != 1 {
		return errors.New("wrong number of arguments")
	}

	isLiquid := ctx.Bool("liquid")

	// The embedded proxy listens on port 3000 (bitcoin) or 3001 (liquid)
	requestPort := "3000"
	if isLiquid {
		requestPort = "3001"
	}
	hex := []byte(ctx.Args().First())

	res, err := http.Post("http://127.0.0.1:"+requestPort+"/tx", "application/string", bytes.NewBuffer(hex))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(string(data))
	}

	if string(data) == "" {
		return errors.New("not successful")
	}
	fmt.Println("\ntxId: " + string(data))
	return nil
}
