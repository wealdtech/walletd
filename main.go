package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/walletd/core"
)

func main() {
	log.SetLevel(log.DebugLevel)

	config, err := core.NewConfig()
	if err != nil {
		panic(err)
	}
	stores, err := core.InitStores(config.Stores)
	if err != nil {
		panic(err)
	}

	rules, err := core.InitRules(config.Rules)
	if err != nil {
		panic(err)
	}

	service, err := NewWalletService(stores, rules)
	if err != nil {
		panic(err)
	}

	if err := service.ServeGRPC(); err != nil {
		panic(err)
	}
}
