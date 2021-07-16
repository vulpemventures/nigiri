package constants

import (
	"errors"
)

const (
	// Datadir key in config json
	Datadir = "datadir"
	// Network key in config json
	Network = "network"
	// Filename key  in config json
	Filename = "nigiri.config.json"
	// AttachLiquid key in config json
	AttachLiquid = "attachLiquid"
	// Version key in config json
	Version = "version"
)

var (
	AvaliableNetworks   = []string{"regtest"}
	NigiriBitcoinImages = []string{
		"ghcr.io/vulpemventures/bitcoin:latest",
		"ghcr.io/vulpemventures/electrs:latest",
		"ghcr.io/vulpemventures/esplora:latest",
		"ghcr.io/vulpemventures/nigiri-chopsticks:latest",
	}
	NigiriLiquidImages = []string{
		"ghcr.io/vulpemventures/elements:latest",
		"ghcr.io/vulpemventures/electrs-liquid:latest",
	}
	NigiriImages = append(NigiriBitcoinImages, NigiriLiquidImages...)
	DefaultEnv   = map[string]interface{}{
		"ports": map[string]map[string]int{
			"bitcoin": {
				"peer":        18432,
				"node":        18433,
				"esplora":     5000,
				"electrs":     3002,
				"electrs_rpc": 51401,
				"chopsticks":  3000,
			},
			"liquid": {
				"peer":        7040,
				"node":        7041,
				"esplora":     5001,
				"electrs":     3012,
				"electrs_rpc": 60401,
				"chopsticks":  3001,
			},
		},
		"urls": map[string]string{
			"bitcoin_esplora": "http://localhost:3000",
			"liquid_esplora":  "http://localhost:3001",
		},
	}

	ErrInvalidNetwork         = errors.New("Network provided is not valid")
	ErrInvalidDatadir         = errors.New("Datadir provided is not valid: it must be an absolute path")
	ErrInvalidServiceName     = errors.New("Service provided is not valid")
	ErrInvalidArgs            = errors.New("Invalid number of args")
	ErrInvalidJSON            = errors.New("JSON environment provided is not valid: missing required fields")
	ErrMalformedJSON          = errors.New("Failed to parse malformed JSON environment")
	ErrEmptyJSON              = errors.New("JSON environment provided is not valid: it must not be empty")
	ErrDatadirNotExisting     = errors.New("Datadir provided is not valid: it must be an existing path")
	ErrNigiriNotRunning       = errors.New("Nigiri is not running")
	ErrNigiriNotExisting      = errors.New("Nigiri does not exists, cannot delete")
	ErrNigiriAlreadyRunning   = errors.New("Nigiri is already running, please stop it first")
	ErrNigiriLiquidNotEnabled = errors.New("Nigiri has been started with no Liquid sidechain.\nPlease stop and restart it using the --liquid flag")
	ErrDockerNotRunning       = errors.New("Nigiri requires the Docker daemon to be running, but it not seems to be started")
)
