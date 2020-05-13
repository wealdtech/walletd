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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/shibukawa/configdir"
)

// Permissions provides information about per-client permissions.
type Permissions struct {
	Certs []*CertificateInfo `json:"certificates"`
}

// CertificateInfo contains information related to client certificates.
type CertificateInfo struct {
	Name  string              `json:"name"`
	Perms []*CertificatePerms `json:"permissions"`
}

// CertificatePerms contains information about the operations allowed by the certificate.
type CertificatePerms struct {
	Path       string   `json:"path"`
	Operations []string `json:"operations"`
}

// FetchPermissions fetches permissions from the JSON configuration file.
func FetchPermissions() (*Permissions, error) {
	configDirs := configdir.New("wealdtech", "walletd")
	configPath := configDirs.QueryFolders(configdir.Global)[0].Path
	path := filepath.Join(configPath, "perms.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	perms := &Permissions{}
	err = json.Unmarshal(data, perms)
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// DumpPerms dumps information about our permissions to stdout.
func DumpPerms(perms *Permissions) {
	for i, certInfo := range perms.Certs {
		if certInfo.Name == "" {
			fmt.Printf("ERROR: certificate %d does not have a name\n", i)
		} else {
			fmt.Printf("Permissions for %q:\n", certInfo.Name)
			for _, perm := range certInfo.Perms {
				if len(perm.Operations) == 1 && perm.Operations[0] == "All" {
					fmt.Printf("\t- accounts matching the path %q can carry out all operations\n", perm.Path)
				} else {
					fmt.Printf("\t- accounts matching the path %q can carry out operations: %s\n", perm.Path, strings.Join(perm.Operations, ", "))
				}
			}
		}
	}
}
