package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var FaucetCmd = &cobra.Command{
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return errors.New("missing address")
		}
		return nil
	},
	Use:     "faucet <address> [amount] [asset]",
	Short:   "Generate and send bitcoin to given address",
	RunE:    faucet,
	PreRunE: faucetChecks,
}

func faucetChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")
	isLiquidService, _ := cmd.Flags().GetBool("liquid")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
		return err
	}
	if len(args) < 1 {
		return constants.ErrInvalidArgs
	}

	if isRunning, err := ctl.IsNigiriRunning(); err != nil {
		return err
	} else if !isRunning {
		return constants.ErrNigiriNotRunning
	}

	if err := ctl.ReadConfigFile(datadir); err != nil {
		return err
	}

	if isLiquidService && isLiquidService != ctl.GetConfigBoolField(constants.AttachLiquid) {
		return constants.ErrNigiriLiquidNotEnabled
	}

	return nil
}

func faucet(cmd *cobra.Command, args []string) error {
	isLiquidService, err := cmd.Flags().GetBool("liquid")
	datadir, _ := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}
	request := map[string]interface{}{
		"address": args[0],
	}
	if len(args) >= 2 {
		amountFloat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %v", err)
		}
		request["amount"] = amountFloat
	}
	if len(args) == 3 {
		request["asset"] = args[2]
	}

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}
	envPath := ctl.GetResourcePath(datadir, "env")
	env, _ := ctl.ReadComposeEnvironment(envPath)
	envPorts := env["ports"].(map[string]map[string]int)
	requestPort := envPorts["bitcoin"]["chopsticks"]
	if isLiquidService {
		requestPort = envPorts["liquid"]["chopsticks"]
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	res, err := http.Post("http://127.0.0.1:"+strconv.Itoa(requestPort)+"/faucet", "application/json", bytes.NewBuffer(payload))
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
		return errors.New("Not Successful")
	}
	fmt.Println("txId: " + dat["txId"])
	return nil
}
