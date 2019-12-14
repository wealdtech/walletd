package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/shibukawa/configdir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	s3 "github.com/wealdtech/go-eth2-wallet-store-s3"
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
)

type config struct {
	Stores []*store `json:"stores"`
}

type store struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Passphrase string `json:"passphrase"`
}

// initConfig initialises the configuration.
// Configuration can come from the configuration file or environment variables.
func initConfig() (*config, error) {
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

	c := &config{}
	return c, viper.Unmarshal(&c)
}

// initStores initialises the stores.
func initStores(c *config) ([]wtypes.Store, error) {
	if len(c.Stores) == 0 {
		log.Warn("No stores configured; using default")
		return initDefaultStores(), nil
	}
	res := make([]wtypes.Store, len(c.Stores))
	for i, store := range c.Stores {
		if store.Name == "" {
			return nil, fmt.Errorf("store %d has no name", i)
		}
		if store.Type == "" {
			return nil, fmt.Errorf("store %d has no type", i)
		}
		switch store.Type {
		case "filesystem":
			res = append(res, filesystem.New(filesystem.WithPassphrase([]byte(store.Passphrase))))
		case "s3":
			s3Store, err := s3.New(s3.WithPassphrase([]byte(store.Passphrase)))
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to access store %d", i))
			}
			res = append(res, s3Store)
		default:
			return nil, fmt.Errorf("store %d has unhandled type %q", i, store.Type)
		}
	}
	return res, nil
}

// initDefaultStores initialises the default stores.
func initDefaultStores() []wtypes.Store {
	res := make([]wtypes.Store, 1)
	res[0] = filesystem.New()
	return res
}
