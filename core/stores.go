package core

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	s3 "github.com/wealdtech/go-eth2-wallet-store-s3"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	wtypes "github.com/wealdtech/go-eth2-wallet-types"
)

// Store defines a store within the configuration
type Store struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Protected  bool   `json:"protected"`
	Passphrase string `json:"passphrase"`
}

// InitStores initialises the stores from a configuration.
func InitStores(c *Config) ([]wtypes.Store, error) {
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
			log.WithField("name", store.Name).Debug("Adding filesystem store")
			res = append(res, filesystem.New(filesystem.WithPassphrase([]byte(store.Passphrase))))
		case "s3":
			log.WithField("name", store.Name).Debug("Adding S3 store")
			s3Store, err := s3.New(s3.WithPassphrase([]byte(store.Passphrase)))
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to access store %d", i))
			}
			res = append(res, s3Store)
		case "scratch":
			log.Debug("Adding scratch store")
			res = append(res, scratch.New())
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
