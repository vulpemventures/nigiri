package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/nigiri/internal/docker"
)

var mint = cli.Command{
	Name:      "mint",
	Usage:     "liquid only: issue and send a given quantity of an asset",
	ArgsUsage: "<address> <amount> [name] [ticker]",
	Action:    mintAction,
}

func mintAction(ctx *cli.Context) error {

	if isRunning, _ := nigiriState.GetBool("running"); !isRunning {
		return errors.New("nigiri is not running")
	}

	if ctx.NArg() < 2 || ctx.NArg() > 5 {
		return errors.New("wrong number of arguments")
	}

	datadir := ctx.String("datadir")
	composePath := getCompose(datadir, true)

	serviceName := "chopsticks-liquid"

	portSlice, err := docker.GetPortsForService(composePath, serviceName)
	if err != nil {
		return err
	}
	mappedPorts := strings.Split(portSlice[0], ":")

	var request struct {
		Address  string `json:"address"`
		Quantity int    `json:"quantity"`
		Name     string `json:"name"`
		Ticker   string `json:"ticker"`
	}
	request.Address = ctx.Args().First()
	request.Quantity, _ = strconv.Atoi(ctx.Args().Get(1))
	if ctx.Args().Len() >= 3 {
		request.Name = ctx.Args().Get(2)
	}
	if ctx.Args().Len() == 4 {
		request.Ticker = ctx.Args().Get(3)
	}

	requestPort := mappedPorts[0]
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	res, err := http.Post("http://127.0.0.1:"+requestPort+"/mint", "application/json", bytes.NewBuffer(payload))
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
		return errors.New("internal error try again")
	}
	if dat["txId"] == "" {
		return errors.New("not successful")
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
