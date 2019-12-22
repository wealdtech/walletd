package core

import (
	"github.com/shibukawa/configdir"
	"github.com/spf13/viper"
)

// Config is the configuration for the daemon.
type Config struct {
	Stores []*Store          `json:"stores"`
	Rules  []*RuleDefinition `json:"rules"`
}

// NewConfig creates a new configuration.
// Configuration can come from the configuration file or environment variables.
func NewConfig() (*Config, error) {
	viper.SetConfigName("config")

	configDirs := configdir.New("wealdtech", "walletd")
	viper.AddConfigPath(configDirs.QueryFolders(configdir.Global)[0].Path)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	viper.SetEnvPrefix("walletd")
	if err := viper.BindEnv("stores"); err != nil {
		return nil, err
	}

	c := &Config{}
	return c, viper.Unmarshal(&c)
}
