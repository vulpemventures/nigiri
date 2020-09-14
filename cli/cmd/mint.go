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

var MintCmd = &cobra.Command{
	Use:     "mint <address> <ammount>",
	Short:   "Generate and send a given quantity of an asset",
	RunE:    mint,
	PreRunE: mintChecks,
}

func mintChecks(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}

	if err := ctl.ParseDatadir(datadir); err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("Invalid number of arguments.\nnigiri mint <address> <ammount> [name] [ticker]")
	}

	if isRunning, err := ctl.IsNigiriRunning(); err != nil {
		return err
	} else if !isRunning {
		return constants.ErrNigiriNotRunning
	}

	if err := ctl.ReadConfigFile(datadir); err != nil {
		return err
	}

	if ctl.GetConfigBoolField(constants.AttachLiquid) != true {
		return constants.ErrNigiriLiquidNotEnabled
	}

	return nil
}

func mint(cmd *cobra.Command, args []string) error {
	datadir, _ := cmd.Flags().GetString("datadir")

	var request struct {
		Address  string `json:"address"`
		Quantity int    `json:"quantity"`
		Name     string `json:"name"`
		Ticker   string `json:"ticker"`
	}
	request.Address = args[0]
	request.Quantity, _ = strconv.Atoi(args[1])
	if len(args) >= 3 {
		request.Name = args[2]
	}
	if len(args) == 4 {
		request.Ticker = args[3]
	}

	ctl, err := controller.NewController()
	if err != nil {
		return err
	}
	envPath := ctl.GetResourcePath(datadir, "env")
	env, _ := ctl.ReadComposeEnvironment(envPath)
	envPorts := env["ports"].(map[string]map[string]int)
	requestPort := envPorts["liquid"]["chopsticks"]

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.Post("http://127.0.0.1:"+strconv.Itoa(requestPort)+"/mint", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var dat map[string]string

	var resp string
	json.Unmarshal([]byte(data), &dat)
	for key, element := range dat {
		resp += key + ": " + element + " "
	}
	fmt.Println(resp[:len(resp)-1])
	return nil
}