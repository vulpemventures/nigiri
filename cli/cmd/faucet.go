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

		if len(args) != 1 {
			return errors.New("Insert receiving address")
		}
		return nil
	},
	Use:     "faucet <bitcoinAddress>",
	Short:   "calls the chopsticks endpoint /faucet to generate and send BTC",
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
	if len(args) != 1 {
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

func faucet(cmd *cobra.Command, address []string) error {
	isLiquidService, err := cmd.Flags().GetBool("liquid")
	datadir, _ := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}
	request := map[string]string{
		"address": address[0],
	}
	ctl, err := controller.NewController()
	if err != nil {
		return err
	}
	envPath := ctl.GetResourcePath(datadir, "env")
	env, _ := ctl.ReadComposeEnvironment(envPath)
	envPorts := env["ports"].(map[string]map[string]int)

	requestPort := envPorts["bitcoin"]["chopsticks"]
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	if isLiquidService {
		requestPort = envPorts["liquid"]["chopsticks"]
	}
	req, err := http.Post("http://localhost:"+strconv.Itoa(requestPort)+"/faucet", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
