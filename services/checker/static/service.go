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

package static

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/util"
)

// StaticChecker checks against a static list.
type StaticChecker struct {
	access map[string][]*path
}

type path struct {
	wallet     *regexp.Regexp
	account    *regexp.Regexp
	operations []string
}

// New creates a new static checker.
func New(ctx context.Context, config *core.Permissions) (*StaticChecker, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "checker.static.New")
	defer span.Finish()

	if config == nil {
		return nil, errors.New("certificate info is required")
	}
	if config.Certs == nil {
		return nil, errors.New("certificates are required")
	}
	if len(config.Certs) == 0 {
		return nil, errors.New("certificate info empty")
	}

	access := make(map[string][]*path, len(config.Certs))
	for _, certificateInfo := range config.Certs {
		if certificateInfo.Name == "" {
			return nil, errors.New("certificate info requires a name")
		}
		if len(certificateInfo.Perms) == 0 {
			return nil, errors.New("certificate info requires at least one permission")
		}
		paths := make([]*path, len(certificateInfo.Perms))
		for i, permissions := range certificateInfo.Perms {
			walletName, accountName, err := util.WalletAndAccountNamesFromPath(permissions.Path)
			if err != nil {
				return nil, fmt.Errorf("invalid account path %s", permissions.Path)
			}
			if walletName == "" {
				return nil, errors.New("wallet cannot be blank")
			}
			walletRegex, err := regexify(walletName)
			if err != nil {
				return nil, fmt.Errorf("invalid wallet regex %s", walletName)
			}
			accountRegex, err := regexify(accountName)
			if err != nil {
				return nil, fmt.Errorf("invalid account regex %s", accountName)
			}
			paths[i] = &path{
				wallet:     walletRegex,
				account:    accountRegex,
				operations: permissions.Operations,
			}
		}
		access[certificateInfo.Name] = paths
	}
	return &StaticChecker{
		access: access,
	}, nil
}

// Check checks the client to see if the account is allowed.
func (c *StaticChecker) Check(ctx context.Context, client string, account string, operation string) bool {
	span, _ := opentracing.StartSpanFromContext(ctx, "checker.static.Check")
	defer span.Finish()

	if client == "" {
		log.Info().Msg("No client certificate name")
		return false
	}
	log := log.With().Str("client", client).Str("account", account).Logger()

	walletName, accountName, err := util.WalletAndAccountNamesFromPath(account)
	if err != nil {
		log.Debug().Err(err).Msg("Invalid path")
		return false
	}
	if walletName == "" {
		log.Debug().Err(err).Msg("Missing wallet name")
		return false
	}
	if accountName == "" {
		log.Debug().Err(err).Msg("Missing account name")
		return false
	}

	paths, exists := c.access[client]
	if !exists {
		log.Debug().Msg("Unknown client")
		return false
	}

	for _, path := range paths {
		if path.wallet.Match([]byte(walletName)) && path.account.Match([]byte(accountName)) {
			for i := range path.operations {
				if path.operations[i] == "All" || path.operations[i] == operation {
					return true
				}
			}
		}
	}
	return false
}

func regexify(name string) (*regexp.Regexp, error) {
	// Empty equates to all.
	if name == "" {
		name = ".*"
	}
	// Anchor if required.
	if !strings.HasPrefix(name, "^") {
		name = fmt.Sprintf("^%s", name)
	}
	if !strings.HasSuffix(name, "$") {
		name = fmt.Sprintf("%s$", name)
	}

	return regexp.Compile(name)

}
