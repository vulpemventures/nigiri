package proxy

// Faucet handles sending funds from the node's wallet
type Faucet struct {
	URL       string
	rpcClient *RPCClient
}

// NewFaucet initializes a faucet and returns it
func NewFaucet(url string, client *RPCClient) *Faucet {
	return &Faucet{url, client}
}

// SendBitcoinTransaction calls the sendtoaddress method of the bitcoin node to the given address with the fractional amount
func (f *Faucet) SendBitcoinTransaction(address string, amount float64) (int, string, error) {
	status, resp, err := HandleRPCRequest(f.rpcClient, "sendtoaddress", []interface{}{address, amount})
	if err != nil {
		return status, "", err
	}

	return status, resp.(string), nil
}

// SendLiquidTransaction calls the sendtoaddress method of the elements node to the given address with the fractional amount of the given asset hash.
// If asset hash is empty will send Liquid Bitcoin
func (f *Faucet) SendLiquidTransaction(address string, amount float64, asset string) (int, string, error) {
	status, resp, err := HandleRPCRequest(f.rpcClient, "sendtoaddress", []interface{}{address, amount, "", "", false, false, 1, "UNSET", false, asset})
	if err != nil {
		return status, "", err
	}

	return status, resp.(string), nil
}

// Fund "matures" the balance by mining blocks if not already mined
// liquid starts with initialfreecoins = 21,000,000 LBTC
func (f *Faucet) Fund(numBlocks int) (int, []string, error) {
	status, resp, err := HandleRPCRequest(f.rpcClient, "getblockcount", nil)
	if err != nil {
		return status, nil, err
	}

	if blockCount := resp.(float64); blockCount <= 0 {
		return f.Mine(numBlocks)
	}

	return 200, nil, nil
}

// Mine will generate blocks to an address of the wallet
func (f *Faucet) Mine(blocks int) (int, []string, error) {
	status, resp, err := HandleRPCRequest(f.rpcClient, "getnewaddress", nil)
	if err != nil {
		return status, nil, err
	}
	address := resp.(string)

	status, resp, err = HandleRPCRequest(f.rpcClient, "generatetoaddress", []interface{}{blocks, address})
	if err != nil {
		return status, nil, err
	}

	blockHashes := []string{}
	for _, b := range resp.([]interface{}) {
		blockHashes = append(blockHashes, b.(string))
	}

	return status, blockHashes, nil
}

// Mint issues a new Liquid asset
func (f *Faucet) Mint(address string, quantity float64) (int, map[string]interface{}, error) {
	status, resp, err := HandleRPCRequest(f.rpcClient, "issueasset", []interface{}{quantity, 0, false})
	if err != nil {
		return status, nil, err
	}
	decodedResp := resp.(map[string]interface{})
	asset := decodedResp["asset"].(string)
	issuanceInput := map[string]interface{}{
		"txid": decodedResp["txid"].(string),
		"vin":  decodedResp["vin"].(float64),
	}

	status, tx, err := f.SendLiquidTransaction(address, quantity, asset)
	if err != nil {
		return status, nil, err
	}

	res := make(map[string]interface{})
	res["asset"] = asset
	res["txId"] = tx
	res["issuance_txin"] = issuanceInput

	return status, res, nil
}
