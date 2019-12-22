package backend

import (
	"github.com/wealdtech/walletd/core"
)

// StaticRuler contains a static list of rules.
type StaticRuler struct {
	rules []*core.Rule
}

// NewStaticruler creates a new static ruler.
func NewStaticRuler(rules []*core.Rule) Ruler {
	return &StaticRuler{
		rules: rules,
	}
}

// Rules fetches matching rules.
func (r *StaticRuler) Rules(request string, account string) []*core.Rule {
	res := make([]*core.Rule, 0)
	for _, rule := range r.rules {
		if rule.Matches(request, account) {
			res = append(res, rule)
		}
	}
	return res
}
