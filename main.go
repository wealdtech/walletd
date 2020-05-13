package main

import (
	"context"
	"flag"
	"os"
	"runtime"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	pprof := ""
	flag.StringVar(&pprof, "pprof", "", "address of a pprof interface for profiling")
	trace := false
	flag.BoolVar(&trace, "trace", false, "provide opentracing stats")
	flag.Parse()

	if pprof != "" {
		go func() {
			// 100 as per https://github.com/golang/go/issues/23401#issuecomment-367029643
			runtime.SetMutexProfileFraction(100)
			if err := http.ListenAndServe(pprof, nil); err != nil {
				log.Warn().Err(err).Msg("Failed to start pprof server")
			}
		}()
	}

	runtime.GOMAXPROCS(runtime.NumCPU() * 8)

	ctx := context.Background()
	if trace {
		tracer, closer, err := InitTracer("walletd")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialise tracer")
		}
		defer closer.Close()
		opentracing.SetGlobalTracer(tracer)
		span := tracer.StartSpan("main")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if err := e2types.InitBLS(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise BLS library")
	}

	// Fetch the configuration.
	config, err := core.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to obtain configuration")
	}

	if showCerts {
		// Need to dump our certificate information.
		core.DumpCerts(config.Server)
		os.Exit(0)
	}

	switch strings.ToLower(config.Verbosity) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "err", "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}

	permissions, err := core.FetchPermissions()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to obtain permissions")
	}
	if showPerms {
		// Need to dump our permission information.
		core.DumpPerms(permissions)
		os.Exit(0)
	}

	// Initialise the keymanager stores.
	stores, err := core.InitStores(ctx, config.Stores)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise stores")
	}

	// Initialise the rules.
	rules, err := core.InitRules(ctx, config.Rules)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise rules")
	}

	// Set up the autounlocker.
	var autounlocker autounlocker.Service
	keysConfig, err := core.FetchKeysConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to obtain keys config")
	}
	if keysConfig != nil {
		autounlocker, err = keys.New(ctx, keysConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialise keys-based autounlocker")
		}
	}

	// Set up the checker.
	checker, err := staticchecker.New(ctx, permissions)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise certificate checker")
	}

	// Initialise the wallet GRPC service.
	service, err := wallet.New(ctx, autounlocker, checker, stores, rules)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create daemon")
	}

	// Start.
	if err := service.ServeGRPC(ctx, config.Server); err != nil {
		log.Fatal().Err(err).Msg("Error running daemon")
	}
}
