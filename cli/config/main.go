package config

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/vulpemventures/nigiri/cli/constants"
)

var vip *viper.Viper

func init() {
	vip = viper.New()
	defaults := viper.New()

	newDefaultConfig(defaults)
	setConfigFromDefaults(vip, defaults)
}

type Config struct{}

func (c *Config) Viper() *viper.Viper {
	return vip
}

func (c *Config) ReadFromFile(path string) error {
	vip.SetConfigFile(filepath.Join(path, constants.Filename))
	return vip.ReadInConfig()
}

func (c *Config) WriteConfig(path string) error {
	vip.SetConfigFile(path)
	return vip.WriteConfig()
}

func (c *Config) GetString(str string) string {
	return vip.GetString(str)
}

func (c *Config) GetBool(str string) bool {
	return vip.GetBool(str)
}

func (c *Config) GetPath() string {
	return getPath()
}

func getPath() string {
	home, _ := homedir.Expand("~")
	return filepath.Join(home, ".nigiri")
}

func newDefaultConfig(v *viper.Viper) {
	v.SetDefault(constants.Datadir, getPath())
	v.SetDefault(constants.Network, "regtest")
	v.SetDefault(constants.AttachLiquid, false)
}

func setConfigFromDefaults(v *viper.Viper, d *viper.Viper) {
	for key, value := range d.AllSettings() {
		v.SetDefault(key, value)
	}
}
