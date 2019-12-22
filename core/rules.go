package core

import (
	"io/ioutil"
	"regexp"

	"github.com/sirupsen/logrus"
)

// RuleDefinition defines a rule.
type RuleDefinition struct {
	Name    string `json:"name"`
	Request string `json:"request"`
	Account string `json:"account"`
	Script  string `json:"script"`
}

// InitRules initialises the rules from a configuration.
func InitRules(defs []*RuleDefinition) ([]*Rule, error) {
	rules := make([]*Rule, 0, len(defs))
	for _, def := range defs {
		rule, err := NewRule(def)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// Rule contains a ready-to-run rule script.
type Rule struct {
	name    string
	request string
	account string
	script  string
}

// Script returns the script for the rule.
func (r *Rule) Script() string {
	return r.script
}

// Matches returns true if this rule matches the path.
func (r *Rule) Matches(request string, path string) bool {
	if r.request != request {
		return false
	}
	res, err := regexp.Match(r.account, []byte(path))
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request": request,
			"path":    path,
			"rule":    r.name,
			"account": r.account,
		}).Warn("Match attempt failed")
		return false
	}
	return res
}

// NewRule creates a new rule from its definition.
func NewRule(def *RuleDefinition) (*Rule, error) {
	contents, err := ioutil.ReadFile(def.Script)
	if err != nil {
		return nil, err
	}

	return &Rule{
		name:    def.Name,
		request: def.Request,
		account: def.Account,
		script:  string(contents),
	}, nil
}
