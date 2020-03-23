package static

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/wealdtech/walletd/core"
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
func New(config []*core.CertificateInfo) (*StaticChecker, error) {
	if config == nil {
		return nil, errors.New("certificate info is required")
	}
	if len(config) == 0 {
		return nil, errors.New("certificate info empty")
	}

	access := make(map[string][]*path, len(config))
	for _, certificateInfo := range config {
		if certificateInfo.Name == "" {
			return nil, errors.New("certificate info requires a name")
		}
		if len(certificateInfo.Permissions) == 0 {
			return nil, errors.New("certificate info requires at least one permission")
		}
		paths := make([]*path, len(certificateInfo.Permissions))
		for i, permissions := range certificateInfo.Permissions {
			if permissions.Path == "" {
				return nil, errors.New("permission path cannot be blank")
			}
			walletName, accountName, err := walletAndAccountNamesFromPath(permissions.Path)
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
func (c *StaticChecker) Check(client string, account string, operation string) bool {
	log := log.WithField("client", client).WithField("account", account)

	walletName, accountName, err := walletAndAccountNamesFromPath(account)
	if err != nil {
		log.WithError(err).Debug("Invalid path")
		return false
	}
	if walletName == "" {
		log.WithError(err).Debug("Missing wallet name")
		return false
	}
	if accountName == "" {
		log.WithError(err).Debug("Missing account name")
		return false
	}

	paths, exists := c.access[client]
	if !exists {
		log.Debug("Unknown client")
		return false
	}

	for _, path := range paths {
		if path.wallet.Match([]byte(walletName)) && path.account.Match([]byte(accountName)) {
			for i := range path.operations {
				if path.operations[i] == operation {
					return true
				}
			}
		}
	}
	return false
}

// walletAndAccountNamesFromPath is a helper that breaks out a path's components.
func walletAndAccountNamesFromPath(path string) (string, string, error) {
	if len(path) == 0 {
		return "", "", errors.New("invalid account format")
	}
	index := strings.Index(path, "/")
	if index == -1 {
		// Just the wallet
		return path, "", nil
	}
	if index == len(path)-1 {
		// Trailing /
		return path[:index], "", nil
	}
	return path[:index], path[index+1:], nil
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
