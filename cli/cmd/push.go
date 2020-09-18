package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vulpemventures/nigiri/cli/constants"
	"github.com/vulpemventures/nigiri/cli/controller"
)

var PushCmd = &cobra.Command{
	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			return errors.New("Missing hex encoded transaction")
		}
		return nil
	},
	Use:     "push <hex>",
	Short:   "Broadcast raw transaction",
	RunE:    push,
	PreRunE: pushChecks,
}

func pushChecks(cmd *cobra.Command, args []string) error {
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

func push(cmd *cobra.Command, args []string) error {
	isLiquidService, err := cmd.Flags().GetBool("liquid")
	datadir, _ := cmd.Flags().GetString("datadir")
	if err != nil {
		return err
	}
	request := args[0]
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
	hex := []byte(request)
	res, err := http.Post("http://127.0.0.1:"+strconv.Itoa(requestPort)+"/tx", "application/string", bytes.NewBuffer(hex))
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
		return errors.New("Not Successful")
	}
	fmt.Println("\ntxId: " + string(data))
	return nil
}
