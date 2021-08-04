package config

import (
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcutil"
)

var (
	DefaultName    = "nigiri.config.json"
	DefaultDatadir = btcutil.AppDataDir("nigiri", false)
	DefaultPath    = filepath.Join(DefaultDatadir, DefaultName)

	InitialState = map[string]string{
		"attachliquid": strconv.FormatBool(false),
		"datadir":      DefaultDatadir,
		"network":      "regtest",
		"ready":        strconv.FormatBool(false),
		"running":      strconv.FormatBool(false),
	}
)
