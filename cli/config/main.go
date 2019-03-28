package config

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	Datadir  = "datadir"
	Network  = "network"
	Filename = "nigiri.config.json"
)

var vip *viper.Viper

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("NIGIRI")
	vip.AutomaticEnv()
	vip.BindEnv("config")

	defaults := viper.New()
	newDefaultConfig(defaults)
	setConfigFromDefaults(vip, defaults)
	vip.SetConfigFile(GetFullPath())
}

func Viper() *viper.Viper {
	return vip
}

func ReadFromFile() error {
	vip.AddConfigPath(GetFullPath())
	return vip.ReadInConfig()
}

func GetPath() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri")
}

func GetFullPath() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri", Filename)
}

func newDefaultConfig(v *viper.Viper) {
	v.SetDefault("datadir", GetPath())
	v.SetDefault("network", "regtest")
	v.SetDefault("version", "0.1.0")
}

func setConfigFromDefaults(v *viper.Viper, d *viper.Viper) {
	for key, value := range d.AllSettings() {
		v.SetDefault(key, value)
	}
}
