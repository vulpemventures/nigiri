package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/vulpemventures/nigiri/internal/proxy/middleware"
)

// Router extends gorilla Router
type Router struct {
	*mux.Router
	Config    Config
	RPCClient *RPCClient
	Faucet    *Faucet
	Registry  *Registry
}

// NewRouter returns a new Router instance
func NewRouter(config Config) *Router {
	router := mux.NewRouter().StrictSlash(true)

	// Create RPC client - forcing calls against the bitcoin/elements default wallet
	rpcClient, _ := NewRPCClient(config.RPCServerURL()+"/wallet/"+config.WalletName(), false, 10)

	r := &Router{router, config, rpcClient, nil, nil}

	// Handle all preflight requests
	r.Router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
	})

	if r.Config.IsLoggerEnabled() {
		r.Use(middleware.Logger)
	}

	r.HandleFunc("/getnewaddress", r.HandleAddressRequest).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/tx", r.HandleBroadcastRequest).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/txs/package", r.HandleSubmitPackageRequest).Methods(http.MethodPost, http.MethodOptions)

	if r.Config.IsFaucetEnabled() {
		r.HandleFunc("/faucet", r.HandleFaucetRequest).Methods(http.MethodPost, http.MethodOptions)
		if config.Chain() == "liquid" {
			r.HandleFunc("/mint", r.HandleMintRequest).Methods(http.MethodPost, http.MethodOptions)
			r.HandleFunc("/registry", r.HandleRegistryRequest).Methods(http.MethodPost, http.MethodOptions)
		}
	}

	// Catch-all proxy to electrs
	r.PathPrefix("/").HandlerFunc(r.HandleElectrsRequest)

	return r
}

// Initialize performs startup tasks like wallet creation and funding
func (r *Router) Initialize() error {
	// From Bitcoin core 0.21 the default wallet "" is not created anymore,
	// and from v31 empty wallet names are rejected outright, so we always
	// create a named wallet (default: "nigiri") if it isn't already loaded.
	err := CreateWalletIfNotExists(r.RPCClient, r.Config.WalletName())
	if err != nil {
		return err
	}
	log.Println("Wallet check completed")

	if r.Config.IsFaucetEnabled() {
		faucet := NewFaucet(r.Config.RPCServerURL(), r.RPCClient)
		r.Faucet = faucet

		if r.Config.Chain() == "liquid" {
			registry, err := NewRegistry(r.Config.RegistryPath())
			if err != nil {
				log.Printf("Warning: could not initialize registry: %v", err)
			}
			r.Registry = registry
		}

		var numBlockToGenerate int = 1
		if r.Config.Chain() == "bitcoin" {
			numBlockToGenerate = 101
		}
		status, blockHashes, err := r.Faucet.Fund(numBlockToGenerate)

		for err != nil && strings.Contains(err.Error(), "Loading") && status == 500 {
			time.Sleep(2 * time.Second)
			status, blockHashes, err = r.Faucet.Fund(numBlockToGenerate)
		}
		if err != nil {
			log.Printf("Warning: Faucet not funded, check the error: %v (status: %d)", err, status)
		}
		if len(blockHashes) > 0 {
			log.Printf("Faucet has been funded by mining %d blocks", len(blockHashes))
		}

		// From Elements core 0.21 if we use initialfreecoins we must rescan the chain
		err = RescanBlockchain(r.RPCClient)
		if err != nil {
			log.Printf("Warning: rescan blockchain failed: %v", err)
		}
		log.Println("Rescan completed")
	}

	return nil
}

// HandleAddressRequest calls the `getnewaddress` and returns the native segwit one
func (r *Router) HandleAddressRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	status, resp, err := HandleRPCRequest(r.RPCClient, "getnewaddress", []interface{}{})
	if err != nil {
		http.Error(res, err.Error(), status)
		return
	}

	json.NewEncoder(res).Encode(map[string]string{"address": resp.(string)})
}

// HandleBroadcastRequest forwards the request to the electrs HTTP server and mines a block if mining is enabled
func (r *Router) HandleBroadcastRequest(res http.ResponseWriter, req *http.Request) {
	r.HandleElectrsRequest(res, req)

	if r.Config.IsMiningEnabled() && r.Faucet != nil {
		status, blockHashes, err := r.Faucet.Mine(1)
		if err != nil {
			log.Printf("Warning: An unexpected error occurred while mining blocks: %v (status: %d)", err, status)
		} else {
			if r.Config.IsLoggerEnabled() {
				log.Printf("Transaction has been confirmed, blocks mined: %d", len(blockHashes))
			}
		}
	}
}

// HandleSubmitPackageRequest handles the submitpackage request
func (r *Router) HandleSubmitPackageRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	decoder := json.NewDecoder(req.Body)
	var txs []string
	err := decoder.Decode(&txs)
	if err != nil {
		http.Error(res, "Malformed Request: missing txs", http.StatusBadRequest)
		return
	}

	status, resp, err := HandleRPCRequest(r.RPCClient, "submitpackage", []interface{}{txs})
	if err != nil {
		http.Error(res, err.Error(), status)
		return
	}

	respMap, ok := resp.(map[string]interface{})
	if !ok {
		http.Error(res, "Malformed Response: expected JSON object", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(res).Encode(respMap); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleFaucetRequest sends funds to the address passed in the request body
func (r *Router) HandleFaucetRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	body := parseRequestBody(req.Body)
	address, ok := body["address"].(string)
	if !ok {
		http.Error(res, "Malformed Request: missing address", http.StatusBadRequest)
		return
	}
	amount, ok := body["amount"].(float64)
	if !ok {
		// the default 100 000 000 satoshis
		amount = 1
	}
	asset, ok := body["asset"].(string)
	if !ok {
		// this means sending bitcoin
		asset = ""
	}

	var status int
	var tx string
	var err error

	if r.Config.Chain() == "liquid" {
		status, tx, err = r.Faucet.SendLiquidTransaction(address, amount, asset)
	} else {
		status, tx, err = r.Faucet.SendBitcoinTransaction(address, amount)
	}
	if err != nil {
		http.Error(res, err.Error(), status)
		return
	}

	if r.Config.IsMiningEnabled() {
		r.Faucet.Mine(1)
	}
	json.NewEncoder(res).Encode(map[string]string{"txId": tx})
}

// HandleMintRequest is a Liquid only endpoint that issues a requested quantity
// of a new asset and sends it to the requested address
func (r *Router) HandleMintRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	body := parseRequestBody(req.Body)
	address, ok := body["address"].(string)
	if !ok {
		http.Error(res, "Malformed Request: missing address", http.StatusBadRequest)
		return
	}
	// NOTICE this is here for backward compatibility. We will deprecate and move to amount
	quantity, qtyOk := body["quantity"].(float64)
	amount, amtOk := body["amount"].(float64)

	if !qtyOk && !amtOk {
		http.Error(res, "Malformed Request: missing amount", http.StatusBadRequest)
		return
	}

	if qtyOk && !amtOk {
		amount = quantity
	}

	status, resp, err := r.Faucet.Mint(address, amount)
	if err != nil {
		http.Error(res, err.Error(), status)
		return
	}

	asset, ok := resp["asset"].(string)
	if !ok {
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}
	issuanceTx, ok := resp["issuance_txin"].(map[string]interface{})
	if !ok {
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}

	name, nameOk := body["name"].(string)
	ticker, tickerOk := body["ticker"].(string)

	if (nameOk && !tickerOk) || (!nameOk && tickerOk) {
		http.Error(res, "Malformed Request: missing name or ticker", http.StatusBadRequest)
		return
	}

	if nameOk && tickerOk && r.Registry != nil {
		contract := map[string]interface{}{
			"name":      name,
			"ticker":    ticker,
			"precision": 8, // we hardcode 8 as precision which is default with issueasset RPC on elements node
		}
		r.Registry.AddEntry(asset, issuanceTx, contract)
	}

	if r.Config.IsMiningEnabled() {
		r.Faucet.Mine(1)
	}

	json.NewEncoder(res).Encode(resp)
}

// HandleRegistryRequest accepts a list of asset ids and returns info retrieved from
// the asset registry about them
func (r *Router) HandleRegistryRequest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	body := parseRequestBody(req.Body)
	assets := body["assets"]
	if assets == nil {
		http.Error(res, "Malformed Request", http.StatusBadRequest)
		return
	}

	if r.Registry == nil {
		http.Error(res, "Registry not initialized", http.StatusInternalServerError)
		return
	}

	entries, err := r.Registry.GetEntries(assets.([]interface{}))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(res).Encode(entries)
}

// transport implements http.RoundTripper to inject registry data into responses
type transport struct {
	registry *Registry
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// check when request path is /asset/{asset_id}
	if t.registry != nil && strings.HasPrefix(req.URL.Path, "/asset/") {
		if s := strings.Split(req.URL.Path, "/"); len(s) == 3 {
			if resp.StatusCode == http.StatusOK {
				// parse response body
				payload, _ := ioutil.ReadAll(resp.Body)
				body := map[string]interface{}{}
				json.Unmarshal(payload, &body)

				// get registry entry for asset
				asset := body["asset_id"].(string)
				entry, _ := t.registry.GetEntry(asset)

				// if entry exists add extra info to response
				if len(entry) > 0 {
					body["name"] = entry["name"]
					body["ticker"] = entry["ticker"]
					body["precision"] = entry["precision"]
					payload, _ = json.Marshal(body)
				}

				newBody := ioutil.NopCloser(bytes.NewReader(payload))
				resp.Body = newBody
				resp.ContentLength = int64(len(payload))
				resp.Header.Set("Content-Length", strconv.Itoa(len(payload)))
			}
		}
	}

	return resp, nil
}

// HandleElectrsRequest forwards every request to the electrs HTTP server
func (r *Router) HandleElectrsRequest(res http.ResponseWriter, req *http.Request) {
	electrsURL := r.Config.ElectrsURL()
	parsedURL, _ := url.Parse(electrsURL)

	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = parsedURL.Host
	req.URL.Host = parsedURL.Host
	req.URL.Scheme = parsedURL.Scheme

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.Transport = &transport{r.Registry}
	proxy.ServeHTTP(res, req)
}

func parseRequestBody(body io.ReadCloser) map[string]interface{} {
	decoder := json.NewDecoder(body)
	var decodedBody map[string]interface{}
	decoder.Decode(&decodedBody)

	return decodedBody
}
