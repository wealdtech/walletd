package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/autounlocker/keys"
	staticchecker "github.com/wealdtech/walletd/services/checker/static"
	"github.com/wealdtech/walletd/services/wallet"
)

func main() {
	if err := e2types.InitBLS(); err != nil {
		log.WithError(err).Fatal("Failed to initialise BLS library")
	}

	showCerts := false
	flag.BoolVar(&showCerts, "show-certs", false, "show server certificates and exit")
	showPerms := false
	flag.BoolVar(&showPerms, "show-perms", false, "show client permissions and exit")
	flag.Parse()

	// Fetch the configuration.
	config, err := core.NewConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to obtain configuration")
	}

	if showCerts {
		// Need to dump our certificate information.
		core.DumpCerts(config.Server)
		os.Exit(0)
	}

	logLevel, err := log.ParseLevel(config.Verbosity)
	if err == nil {
		log.SetLevel(logLevel)
	}

	permissions, err := core.FetchPermissions()
	if err != nil {
		log.WithError(err).Fatal("Failed to obtain permissions")
	}
	if showPerms {
		// Need to dump our permission information.
		core.DumpPerms(permissions)
		os.Exit(0)
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

	// Set up the autounlocker.
	var autounlocker autounlocker.Service
	keysConfig, err := core.FetchKeysConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to obtain keys config")
	}
	if keysConfig != nil {
		autounlocker, err = keys.New(keysConfig)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialise keys-based autounlocker")
		}
	}

	// Set up the checker.
	checker, err := staticchecker.New(permissions)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise certificate checker")
	}

	// Initialise the wallet GRPC service.
	service, err := wallet.New(autounlocker, checker, stores, rules)
	if err != nil {
		log.WithError(err).Fatal("Failed to create daemon")
	}

	// Start.
	if err := service.ServeGRPC(config.Server); err != nil {
		log.WithError(err).Fatal("Error running daemon")
	}
}
