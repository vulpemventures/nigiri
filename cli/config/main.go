package config

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	Datadir      = "datadir"
	Network      = "network"
	Filename     = "nigiri.config.json"
	AttachLiquid = "attachLiquid"
	Version      = "version"
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
	// vip.SetConfigFile(GetFullPath())
}

func Viper() *viper.Viper {
	return vip
}

func ReadFromFile(path string) error {
	vip.SetConfigFile(filepath.Join(path, Filename))
	return vip.ReadInConfig()
}

func WriteConfig(path string) error {
	vip.SetConfigFile(path)
	return vip.WriteConfig()
}

func GetString(str string) string {
	return vip.GetString(str)
}

func GetBool(str string) bool {
	return vip.GetBool(str)
}

func GetPath() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri")
}

func newDefaultConfig(v *viper.Viper) {
	v.SetDefault(Datadir, GetPath())
	v.SetDefault(Network, "regtest")
	v.SetDefault(AttachLiquid, false)
	v.SetDefault(Version, "0.1.0")
}

func setConfigFromDefaults(v *viper.Viper, d *viper.Viper) {
	for key, value := range d.AllSettings() {
		v.SetDefault(key, value)
	}
}
