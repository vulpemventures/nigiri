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
	hex, err := getNewSignedTransaction(bitcoin)
	if err != nil {
		t.Fatal(err)
	}

	if err := testCommand("push", hex, bitcoin); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func TestPushLiquidTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	testStart(t, liquid)
	hex, err := getNewSignedTransaction(liquid)
	if err != nil {
		t.Fatal(err)
	}

	if err := testCommand("push", hex, liquid); err != nil {
		t.Fatal(err)
	}

	testDelete(t)
}

func getNewSignedTransaction(isLiquid bool) (string, error) {
	var hexFinal string
	txId, vout, amount, asset, err := listUnspent(isLiquid)
	if err != nil {
		return "", err
	}
	hex, err := createRawTransaction(txId, vout, amount, asset, isLiquid)
	if err != nil {
		return "", err
	}
	hexFinal, err = signRawTransaction(hex, isLiquid)
	if err != nil {
		return "", err
	}
	return hexFinal, nil
}

func listUnspent(isLiquid bool) (string, string, string, string, error) {
	bashCmd, err := execCommand([]string{"listunspent"}, isLiquid)
	if err != nil {
		return "", "", "", "", err
	}
	type unspent map[string]interface{}
	var unspentList []unspent
	if err := json.Unmarshal([]byte(bashCmd), &unspentList); err != nil {
		return "", "", "", "", errors.New("Internal error. Try again.")
	}
	txId := fmt.Sprintf("%v", unspentList[0]["txid"])
	vout := fmt.Sprintf("%v", unspentList[0]["vout"])
	amount := fmt.Sprintf("%v", unspentList[0]["amount"])

	if isLiquid {
		asset := fmt.Sprintf("%v", unspentList[0]["asset"])
		return txId, vout, amount, asset, nil
	}

	return txId, vout, amount, "", nil
}

func createRawTransaction(txId string, vout string, amount string, asset string, isLiquid bool) (string, error) {
	var bashCmd []byte
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
	sendAdress, err := execCommand([]string{"getnewaddress"}, isLiquid)
	if err != nil {
		return "", err
	}
	outputs := make(map[string]interface{})
	if !isLiquid {
		outputs[string(sendAdress)], _ = strconv.ParseFloat(amountSend, 64)
		outputsJson, err := json.Marshal(outputs)
		if err != nil {
			return "", err
		}
		bashCmd, err = execCommand([]string{"createrawtransaction", string(inputsJson), string(outputsJson)}, bitcoin)
		if err != nil {
			return "", err
		}
	}
	if isLiquid {
		addressInfoJson, err := execCommand([]string{"getaddressinfo", string(sendAdress)}, isLiquid)
		if err != nil {
			return "", err
		}
		var adressInfo map[string]interface{}
		if err := json.Unmarshal([]byte(addressInfoJson), &adressInfo); err != nil {
			return "", errors.New("Internal error. Try again.")
		}
		sendAdressUnc := fmt.Sprintf("%v", adressInfo["unconfidential"])
		outputs[sendAdressUnc], _ = strconv.ParseFloat(amountSend, 64)
		outputs["fee"] = fee
		outputsJson, err := json.Marshal(outputs)
		if err != nil {
			return "", err
		}
		output_assets := make(map[string]interface{})
		output_assets[sendAdressUnc] = asset
		output_assets["fee"] = asset
		output_assetsJson, err := json.Marshal(output_assets)
		if err != nil {
			return "", err
		}
		bashCmd, err = execCommand([]string{"createrawtransaction", string(inputsJson), string(outputsJson), "0", "false", string(output_assetsJson)}, isLiquid)
		if err != nil {
			return "", err
		}
	}
	return strings.Fields(string(bashCmd))[0], nil
}

func signRawTransaction(hex string, isLiquid bool) (string, error) {
	bashCmd, err := execCommand([]string{"signrawtransactionwithwallet", hex}, isLiquid)
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

func execCommand(args []string, isLiquid bool) ([]byte, error) {
	rpcArgs := []string{"exec", "bitcoin", "bitcoin-cli", "-datadir=config"}
	if isLiquid {
		rpcArgs = []string{"exec", "liquid", "elements-cli", "-datadir=config"}
	}
	cmdArgs := append(rpcArgs, args...)
	bashCmd, err := exec.Command("docker", cmdArgs...).Output()
	if err != nil {
		return nil, err
	}
	return bashCmd, nil
}
