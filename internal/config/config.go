package config

import (
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcutil"
)

var (
	DefaultName    = "nigiri.config.json"
	DefaultCompose = "docker-compose.yml"
	SignetCompose  = "docker-compose.signet.yml"

	DefaultDatadir = btcutil.AppDataDir("nigiri", false)
	DefaultPath    = filepath.Join(DefaultDatadir, DefaultName)

	InitialState = map[string]string{
		"network": "regtest",
		"ready":   strconv.FormatBool(false),
		"running": strconv.FormatBool(false),
	}
)
