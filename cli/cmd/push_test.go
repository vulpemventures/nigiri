package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestPushBitcoinTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, bitcoin)
	hex, err := getNewSignedTransaction()
	if err != nil {
		t.Fatal(err)
	}

	if err := testCommand("push", hex, bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func getNewSignedTransaction() (string, error) {
	txId, vout, amount, err := listUnspent()
	if err != nil {
		return "", err
	}
	hex, err := createRawTransaction(txId, vout, amount)
	if err != nil {
		return "", err
	}
	hexFinal, err := signRawTransaction(hex)
	if err != nil {
		return "", err
	}
	return hexFinal, nil
}

func listUnspent() (string, string, string, error) {
	bashCmd, err := execCommand([]string{"listunspent"}, bitcoin)
	if err != nil {
		return "", "", "", err
	}
	type unspent map[string]interface{}
	var unspentList []unspent
	if err := json.Unmarshal([]byte(bashCmd), &unspentList); err != nil {
		return "", "", "", errors.New("Internal error. Try again.")
	}
	txId := fmt.Sprintf("%v", unspentList[0]["txid"])
	vout := fmt.Sprintf("%v", unspentList[0]["vout"])
	amount := fmt.Sprintf("%v", unspentList[0]["amount"])

	return txId, vout, amount, nil
}

func createRawTransaction(txId string, vout string, amount string) (string, error) {

	type inputs map[string]interface{}
	var inputsList [1]inputs
	inputsList[0] = make(inputs, 2)
	inputsList[0]["txid"] = txId
	inputsList[0]["vout"], _ = strconv.Atoi(vout)
	inputsJson, err := json.Marshal(inputsList)
	if err != nil {
		return "", err
	}

	fee := 0.00001
	amountInt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return "", err
	}
	amountSend := fmt.Sprintf("%.6f", amountInt-fee)
	sendAdress, err := execCommand([]string{"getnewaddress"}, bitcoin)
	if err != nil {
		return "", err
	}

	type outputs map[string]interface{}
	var outputsList [1]outputs
	outputsList[0] = make(outputs, 1)
	outputsList[0][string(sendAdress)], _ = strconv.ParseFloat(amountSend, 64)
	outputsJson, err := json.Marshal(outputsList)
	if err != nil {
		return "", err
	}

	bashCmd, err := execCommand([]string{"createrawtransaction", string(inputsJson), string(outputsJson)}, bitcoin)
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
