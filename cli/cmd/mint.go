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
	Use:     "mint <address> <ammount> [name] [ticker]",
	Short:   "Liquid only: Issue and send a given quantity of an asset",
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
		return errors.New("missing required arguments")
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
	res, err := http.Post("http://127.0.0.1:"+strconv.Itoa(requestPort)+"/mint", "application/json", bytes.NewBuffer(payload))
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
	var dat map[string]interface{}
	var resp string
	if err := json.Unmarshal([]byte(data), &dat); err != nil {
		return errors.New("Internal error. Try again.")
	}
	if dat["txId"] == "" {
		return errors.New("Not Successful")
	}
	for key, element := range dat {
		if key == "issuance_txin" {
			myMap := element.(map[string]interface{})
			resp += key + ":\n"
			for key2, element2 := range myMap {
				resp += "  "
				resp += key2 + ": " + fmt.Sprintf("%v", element2) + "\n"
			}
			continue
		}
		resp += key + ": " + fmt.Sprintf("%v", element) + "\n"
	}
	fmt.Println(resp[:len(resp)-1])
	return nil
}
