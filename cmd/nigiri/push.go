package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/docker"
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

	isRunning, err := getBoolFromState("running")
	if err != nil {
		return err
	}

	if !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() != 1 {
		return errors.New("wrong number of arguments")
	}

	isLiquid := ctx.Bool("liquid")
	composePath := getCompose(isLiquid)

	var serviceName string = "chopsticks"
	if isLiquid {
		serviceName = "chopsticks-liquid"
	}

	portSlice, err := docker.GetPortsForService(composePath, serviceName)
	if err != nil {
		return err
	}
	mappedPorts := strings.Split(portSlice[0], ":")
	requestPort := mappedPorts[0]
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
