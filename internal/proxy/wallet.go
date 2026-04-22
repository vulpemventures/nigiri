package proxy

import (
	"fmt"
	"net/http"
)

// CreateWalletIfNotExists checks if a wallet named walletName is loaded and
// creates it if not. Bitcoin Core v31+ rejects empty wallet names, so callers
// must pass a non-empty name.
func CreateWalletIfNotExists(client *RPCClient, walletName string) error {
	status, resp, err := HandleRPCRequest(client, "listwallets", []interface{}{})
	if err != nil {
		return fmt.Errorf("could not list wallets: %w", err)
	}

	loadedWallets, ok := resp.([]interface{})
	if !ok {
		return fmt.Errorf("could not list wallets: unexpected response type")
	}

	if status != http.StatusOK {
		return nil
	}

	// If our wallet is already loaded, nothing to do.
	for _, w := range loadedWallets {
		if name, ok := w.(string); ok && name == walletName {
			return nil
		}
	}

	_, _, err = HandleRPCRequest(client, "createwallet", []interface{}{walletName})
	if err != nil {
		return fmt.Errorf("could not create wallet %q: %w", walletName, err)
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
