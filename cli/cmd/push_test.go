package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestPushBitcoinTransaction(t *testing.T) {
	testStart(t, bitcoin)
	hex, err := getNewSignTransaction()
	if err != nil {
		t.Fatal(err)
	}

	if err := testCommand("push", hex, bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func getNewSignTransaction() (string, error) {
	txId, address, _ := listUnspent()
	hex, err := createRawTransaction(txId, address)
	if err != nil {
		return "", err
	}
	hexFinal, err := signRawTransaction(hex)
	if err != nil {
		return "", err
	}
	return hexFinal, nil
}

func listUnspent() (string, string, error) {
	bashCmd, err := execCommand([]string{"listunspent"}, bitcoin)
	if err != nil {
		return "", "", err
	}
	type unspent map[string]interface{}
	var unspentList []unspent
	if err := json.Unmarshal([]byte(bashCmd), &unspentList); err != nil {
		return "", "", errors.New("Internal error. Try again.")
	}
	txId := fmt.Sprintf("%v", unspentList[0]["txid"])
	address := fmt.Sprintf("%v", unspentList[0]["address"])

	return txId, address, nil
}

func createRawTransaction(txId string, address string) (string, error) {
	inputs := `[{"txid" : "` + txId + `", "vout" : 0}]`
	outputs := `{"` + address + `" : 49.9999}`

	bashCmd, err := execCommand([]string{"createrawtransaction", inputs, outputs}, bitcoin)
	if err != nil {
		return "", err
	}
	return strings.Fields(string(bashCmd))[0], nil
}

func signRawTransaction(hex string) (string, error) {
	bashCmd, err := execCommand([]string{"signrawtransactionwithwallet", hex}, bitcoin)
	if err != nil {
		return "", err
	}
	var hexJson map[string]interface{}
	if err := json.Unmarshal([]byte(bashCmd), &hexJson); err != nil {
		return "", errors.New("Internal error. Try again.")
	}
	hex = fmt.Sprintf("%v", hexJson["hex"])

	return hex, nil
}

func execCommand(args []string, liquid bool) ([]byte, error) {
	rpcArgs := []string{"exec", "bitcoin", "bitcoin-cli", "-datadir=config"}
	if liquid {
		rpcArgs = []string{"exec", "liquid", "elements-cli", "-datadir=config"}
	}
	cmdArgs := append(rpcArgs, args...)
	bashCmd, err := exec.Command("docker", cmdArgs...).Output()
	if err != nil {
		return nil, err
	}
	return bashCmd, nil
}
