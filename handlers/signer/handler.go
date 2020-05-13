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
	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/checker"
	"github.com/wealdtech/walletd/services/fetcher"
	"github.com/wealdtech/walletd/services/ruler"
)

// Handler is the signer handler.
type Handler struct {
	checker      checker.Service
	fetcher      fetcher.Service
	ruler        ruler.Service
	autounlocker autounlocker.Service
}

// New creates a new signer handler.
func New(autounlocker autounlocker.Service, checker checker.Service, fetcher fetcher.Service, ruler ruler.Service) *Handler {
	return &Handler{
		autounlocker: autounlocker,
		checker:      checker,
		fetcher:      fetcher,
		ruler:        ruler,
	}
}
