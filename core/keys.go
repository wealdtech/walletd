package core

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/shibukawa/configdir"
)

// KeysConfig provides information about keys for automatic unlocking.
type KeysConfig struct {
	Keys []string `json:"keys"`
}

// FetchKeysConfig fetches keys from the JSON configuration file.
func FetchKeysConfig() (*KeysConfig, error) {
	configDirs := configdir.New("wealdtech", "walletd")
	configPath := configDirs.QueryFolders(configdir.Global)[0].Path
	path := filepath.Join(configPath, "keys.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	keys := &KeysConfig{}
	err = json.Unmarshal(data, keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}
