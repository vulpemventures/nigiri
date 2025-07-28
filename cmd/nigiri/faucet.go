package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/config"
	"github.com/vulpemventures/nigiri/internal/docker"
)

var faucet = cli.Command{
	Name:      "faucet",
	Usage:     "generate and send bitcoin to given address",
	ArgsUsage: "<address> [amount] [asset]",
	Action:    faucetAction,
	Flags: []cli.Flag{
		&liquidFlag,
		&arkFlag,
	},
}

func faucetAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() < 1 || ctx.NArg() > 3 {
		return errors.New("wrong number of arguments")
	}

	isLiquid := ctx.Bool("liquid")
	datadir := ctx.String("datadir")
	composePath := filepath.Join(datadir, config.DefaultCompose)

	var serviceName string = "chopsticks"
	if isLiquid {
		serviceName = "chopsticks-liquid"
	}

	// Get the port for the service
	dockerClient := docker.NewDefaultClient()
	portSlice, err := dockerClient.GetPortsForService(composePath, serviceName)
	if err != nil {
		return err
	}
	mappedPorts := strings.Split(portSlice[0], ":")

	network, err := nigiriState.GetString("network")
	if err != nil {
		return err
	}

	isArk := ctx.Bool("ark")
	if isArk {
		amount := 1.0
		if ctx.Args().Len() >= 2 {
			amount, err = strconv.ParseFloat(ctx.Args().Get(1), 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %v", err)
			}
		}

		return faucetArk(ctx.Args().First(), amount)
	}

	address := ctx.Args().First()
	if address == "cln" {
		jsonOut, err := outputCommand("docker", "exec", "cln", "lightning-cli", "--network="+network, "newaddr")
		if err != nil {
			return err
		}

		address, err = getValueByKey(jsonOut, "bech32")
		if err != nil {
			return err
		}
	}

	if address == "lnd" {
		jsonOut, err := outputCommand("docker", "exec", "lnd", "lncli", "--network="+network, "newaddress", "p2wkh")
		if err != nil {
			return err
		}

		address, err = getValueByKey(jsonOut, "address")
		if err != nil {
			return err
		}
	}

	request := map[string]interface{}{
		"address": address,
	}

	if ctx.Args().Len() >= 2 {
		amountFloat, err := strconv.ParseFloat(ctx.Args().Get(1), 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %v", err)
		}
		request["amount"] = amountFloat
	}

	if isLiquid && ctx.Args().Len() == 3 {
		request["asset"] = ctx.Args().Get(2)
	}

	requestPort := mappedPorts[0]
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	res, err := http.Post("http://127.0.0.1:"+requestPort+"/faucet", "application/json", bytes.NewBuffer(payload))
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

	var dat map[string]string
	if err := json.Unmarshal([]byte(data), &dat); err != nil {
		return errors.New("internal error, please try again")
	}
	if dat["txId"] == "" {
		return errors.New("not successful")
	}
	fmt.Println("txId: " + dat["txId"])

	return nil
}

func getValueByKey(JSONobject []byte, key string) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal(JSONobject, &data)
	if err != nil {
		return "", err
	}
	return data[key].(string), nil
}

func outputCommand(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("name: %v, args: %v, err: %v", name, arg, err.Error())
	}
	return b, nil
}

func faucetArk(address string, amount float64) error {
	amount = math.Round(amount * 100000000) // convert to satoshis
	bashCmd := exec.Command("docker", "exec", "-t", "ark", "ark", "balance")
	output, err := bashCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Parse the JSON response to extract offchain balance
	var balanceData map[string]interface{}
	if err := json.Unmarshal(output, &balanceData); err != nil {
		return fmt.Errorf("failed to parse balance JSON: %w", err)
	}

	// Extract offchain balance
	offchainBalance, ok := balanceData["offchain_balance"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to get offchain_balance from response")
	}

	total, ok := offchainBalance["total"].(float64)
	if !ok {
		return fmt.Errorf("failed to get total from offchain_balance")
	}

	if total < amount {
		bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "wallet", "address")
		output, err := bashCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get wallet address: %w", err)
		}
		address := strings.TrimSpace(string(output))

		// Fund the address using nigiri faucet
		bashCmd = exec.Command("nigiri", "faucet", address)
		if err := bashCmd.Run(); err != nil {
			return fmt.Errorf("failed to fund wallet address: %w", err)
		}

		bashCmd = exec.Command("docker", "exec", "-t", "ark", "arkd", "note", "--amount", fmt.Sprintf("%d", int64(amount)))
		output, err = bashCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to create arkd note: %w", err)
		}
		noteStr := strings.TrimSpace(string(output))

		bashCmd = exec.Command("docker", "exec", "-t", "ark", "ark", "redeem-notes", "-n", noteStr, "--password", "secret")
		if err := bashCmd.Run(); err != nil {
			return fmt.Errorf("failed to redeem note: %w", err)
		}
	}

	bashCmd = exec.Command("docker", "exec", "-t", "ark", "ark", "send", "--to", address, "--amount", fmt.Sprintf("%d", int64(amount)), "--password", "secret")
	output, err = bashCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to send ark: %w", err)
	}
	fmt.Println(string(output))

	return nil
}
