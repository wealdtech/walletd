package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/walletd"
)

func main() {
	// Fetch the configuration.
	config, err := core.NewConfig()
	if err != nil {
		panic(err)
	}

	logLevel, err := log.ParseLevel(config.Verbosity)
	if err == nil {
		log.SetLevel(logLevel)
	}

	// Initialise the keymanager stores.
	stores, err := core.InitStores(config.Stores)
	if err != nil {
		panic(err)
	}

	// Initialise the rules.
	rules, err := core.InitRules(config.Rules)
	if err != nil {
		panic(err)
	}

	// Initialise the wallet GRPC service.
	service, err := walletd.New(stores, rules)
	if err != nil {
		panic(err)
	}

	// Start.
	if err := service.ServeGRPC(config.Server); err != nil {
		panic(err)
	}
}
