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

package dummy

import (
	"context"
	"strings"
)

// DummyChecker returns true all clients and accounts except those that start with "Deny".  Only for testing.
type DummyChecker struct{}

// New creates a new dummy checker.
func New() (*DummyChecker, error) {
	return &DummyChecker{}, nil
}

// Check returns true.
func (c *DummyChecker) Check(ctx context.Context, client string, account string, operation string) bool {
	return !(strings.HasPrefix(client, "Deny") || strings.Contains(account, "/Deny"))
}
