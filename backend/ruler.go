package backend

import (
	"github.com/wealdtech/walletd/core"
)

// Ruler provides rules matching a path that must succeed before operations can proceed.
type Ruler interface {
	Rules(request string, account string) []*core.Rule
}
