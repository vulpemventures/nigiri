package proxy

import (
	"fmt"
	"net/http"
)

// CreateWalletIfNotExists checks if wallet exists and creates one if needed
func CreateWalletIfNotExists(client *RPCClient) error {
	status, resp, err := HandleRPCRequest(client, "listwallets", []interface{}{})
	if err != nil {
		return fmt.Errorf("could not list wallets: %w", err)
	}

	numOfWallets, ok := resp.([]interface{})
	if !ok {
		return fmt.Errorf("could not list wallets: unexpected response type")
	}

	if status == http.StatusOK && len(numOfWallets) == 0 {
		_, _, err = HandleRPCRequest(client, "createwallet", []interface{}{""})
		if err != nil {
			return fmt.Errorf("could not create wallet: %w", err)
		}
	}

	return nil
}

// RescanBlockchain triggers a rescan of the blockchain
func RescanBlockchain(client *RPCClient) error {
	_, _, err := HandleRPCRequest(client, "rescanblockchain", []interface{}{})
	if err != nil {
		return fmt.Errorf("could not rescan: %w", err)
	}

	return nil
}
