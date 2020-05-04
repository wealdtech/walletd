package main

import (
	"context"
	"flag"
	"os"
	"runtime"

	"net/http"
	_ "net/http/pprof"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	"github.com/wealdtech/walletd/core"
	"github.com/wealdtech/walletd/services/autounlocker"
	"github.com/wealdtech/walletd/services/autounlocker/keys"
	staticchecker "github.com/wealdtech/walletd/services/checker/static"
	"github.com/wealdtech/walletd/services/wallet"
)

func main() {
	showCerts := false
	flag.BoolVar(&showCerts, "show-certs", false, "show server certificates and exit")
	showPerms := false
	flag.BoolVar(&showPerms, "show-perms", false, "show client permissions and exit")
	pprof := false
	flag.BoolVar(&pprof, "pprof", false, "add a pprof interface for profiling")
	trace := false
	flag.BoolVar(&trace, "trace", false, "provide opentracing stats")
	flag.Parse()

	if pprof {
		go func() {
			runtime.SetMutexProfileFraction(1)
			//if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			if err := http.ListenAndServe("0.0.0.0:12333", nil); err != nil {
				log.WithError(err).Warn("Failed to start pprof server")
			}
		}()
	}

	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	ctx := context.Background()
	if trace {
		tracer, closer, err := InitTracer("walletd")
		if err != nil {
			log.WithError(err).Fatal("Failed to initialise tracer")
		}
		defer closer.Close()
		opentracing.SetGlobalTracer(tracer)
		span := tracer.StartSpan("main")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if err := e2types.InitBLS(); err != nil {
		log.WithError(err).Fatal("Failed to initialise BLS library")
	}

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
	stores, err := core.InitStores(ctx, config.Stores)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise stores")
	}

	// Initialise the rules.
	rules, err := core.InitRules(ctx, config.Rules)
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
		autounlocker, err = keys.New(ctx, keysConfig)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialise keys-based autounlocker")
		}
	}

	// Set up the checker.
	checker, err := staticchecker.New(ctx, permissions)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialise certificate checker")
	}

	// Initialise the wallet GRPC service.
	service, err := wallet.New(ctx, autounlocker, checker, stores, rules)
	if err != nil {
		log.WithError(err).Fatal("Failed to create daemon")
	}

	// Start.
	if err := service.ServeGRPC(ctx, config.Server); err != nil {
		log.WithError(err).Fatal("Error running daemon")
	}
}
