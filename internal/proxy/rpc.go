package proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// RPCClient represents a JSON RPC client (over HTTP(s)).
type RPCClient struct {
	serverAddr string
	httpClient *http.Client
	timeout    int
}

// rpcRequest represent a RCP request
type rpcRequest struct {
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int64       `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
}

// RPCResponse holds the parsed JSON-RPC response
type RPCResponse struct {
	Id     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Err    interface{}     `json:"error"`
}

// NewRPCClient creates a new RPC client
func NewRPCClient(url string, useSSL bool, timeout int) (*RPCClient, error) {
	var httpClient *http.Client

	if useSSL {
		t := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = &http.Client{Transport: t}
	} else {
		httpClient = &http.Client{}
	}

	c := &RPCClient{
		serverAddr: url,
		httpClient: httpClient,
		timeout:    timeout,
	}

	return c, nil
}

// Call prepares and executes the RPC request
func (c *RPCClient) Call(method string, params interface{}) (int, RPCResponse, error) {
	status := http.StatusInternalServerError
	var rr RPCResponse

	connectTimer := time.NewTimer(time.Duration(c.timeout) * time.Second)
	rpcR := rpcRequest{method, params, time.Now().UnixNano(), "1.0"}
	payloadBuffer := &bytes.Buffer{}
	if err := json.NewEncoder(payloadBuffer).Encode(rpcR); err != nil {
		return status, rr, err
	}

	req, err := http.NewRequest("POST", c.serverAddr, payloadBuffer)
	if err != nil {
		return status, rr, err
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Accept", "application/json")

	resp, err := c.doTimeoutRequest(connectTimer, req)
	if err != nil {
		return status, rr, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return status, rr, err
	}

	if resp.StatusCode != http.StatusOK {
		out := map[string]map[string]interface{}{}
		json.Unmarshal(data, &out)
		errMsg := "unknown error"
		if errMap, ok := out["error"]; ok {
			if msg, ok := errMap["message"].(string); ok {
				errMsg = msg
			}
		}
		return resp.StatusCode, rr, fmt.Errorf("Method %s failed with error: %s", method, errMsg)
	}

	if err = json.Unmarshal(data, &rr); err != nil {
		return status, rr, err
	}

	return resp.StatusCode, rr, nil
}

// doTimeoutRequest processes a HTTP request with timeout
func (c *RPCClient) doTimeoutRequest(timer *time.Timer, req *http.Request) (*http.Response, error) {
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := c.httpClient.Do(req)
		done <- result{resp, err}
	}()
	// Wait for the read or the timeout
	select {
	case r := <-done:
		return r.resp, r.err
	case <-timer.C:
		return nil, fmt.Errorf("Timeout reading data from server")
	}
}

// HandleRPCRequest calls a JSONRPC method and decodes the JSON body as response
func HandleRPCRequest(client *RPCClient, method string, params []interface{}) (int, interface{}, error) {
	status, resp, err := client.Call(method, params)
	if err != nil {
		return status, "", err
	}
	var out interface{}
	err = json.Unmarshal(resp.Result, &out)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}

	return status, out, nil
}
