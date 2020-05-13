// Copyright © 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
