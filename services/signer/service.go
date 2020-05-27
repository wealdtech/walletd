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

package signer

import (
	"errors"

	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Service is the signer handler.
type Service struct {
	checker      checker.Service
	fetcher      fetcher.Service
	ruler        ruler.Service
	autounlocker autounlocker.Service
}

// New creates a new signer handler.
func New(unlocker autounlocker.Service, checker checker.Service, fetcher fetcher.Service, ruler ruler.Service) (*Service, error) {
	if unlocker == nil {
		return nil, errors.New("no unlocker provided")
	}
	if checker == nil {
		return nil, errors.New("no checker provided")
	}
	if fetcher == nil {
		return nil, errors.New("no fetcher provided")
	}
	if ruler == nil {
		return nil, errors.New("no ruler provided")
	}

	return &Service{
		autounlocker: unlocker,
		checker:      checker,
		fetcher:      fetcher,
		ruler:        ruler,
	}, nil
}
