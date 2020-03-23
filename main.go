package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/walletd/core"
	staticchecker "github.com/wealdtech/walletd/services/checker/static"
	"github.com/wealdtech/walletd/services/wallet"
)

func main() {
	// Fetch the configuration.
	config, err := core.NewConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to obtain configuration")
	}

	logLevel, err := log.ParseLevel(config.Verbosity)
	if err == nil {
		log.SetLevel(logLevel)
	}

	// Initialise the keymanager stores.
	stores, err := core.InitStores(config.Stores)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise stores")
	}

	// Initialise the rules.
	rules, err := core.InitRules(config.Rules)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise rules")
	}

	// Set up the checker.
	checker, err := staticchecker.New(config.Certs)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise certificate checker")
	}

	// Initialise the wallet GRPC service.
	service, err := wallet.New(checker, stores, rules)
	if err != nil {
		log.WithError(err).Fatal("Failed to create daemon")
	}

	// Start.
	if err := service.ServeGRPC(config.Server); err != nil {
		log.WithError(err).Fatal("Error running daemon")
	}
}
