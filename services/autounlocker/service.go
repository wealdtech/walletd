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

package autounlocker

import (
	"context"

	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// Service provides an interface to automatically unlock accounts when required.
type Service interface {
	Unlock(context.Context, e2wtypes.Wallet, e2wtypes.Account) (bool, error)
}
