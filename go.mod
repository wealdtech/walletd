module github.com/wealdtech/walletd

go 1.13

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aws/aws-sdk-go v1.30.11 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prologic/bitcask v0.3.5
	github.com/prysmaticlabs/go-ssz v0.0.0-20200101200214-e24db4d9e963
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0
	github.com/sirupsen/logrus v1.5.0
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.5.1
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/wealdtech/eth2-signer-api v1.3.0
	github.com/wealdtech/go-bytesutil v1.1.1
	github.com/wealdtech/go-eth2-types/v2 v2.3.1
	github.com/wealdtech/go-eth2-wallet v1.9.3
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.0.0
	github.com/wealdtech/go-eth2-wallet-hd/v2 v2.0.1
	github.com/wealdtech/go-eth2-wallet-nd/v2 v2.0.1
	github.com/wealdtech/go-eth2-wallet-store-filesystem v1.7.2
	github.com/wealdtech/go-eth2-wallet-store-s3 v1.6.2
	github.com/wealdtech/go-eth2-wallet-store-scratch v1.3.3
	github.com/wealdtech/go-eth2-wallet-types/v2 v2.0.2
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	go.uber.org/atomic v1.6.0 // indirect
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd // indirect
	google.golang.org/genproto v0.0.0-20200420144010-e5e8543f8aeb // indirect
	google.golang.org/grpc v1.29.0
	gopkg.in/ini.v1 v1.55.0 // indirect
)

replace github.com/wealdtech/go-eth2-wallet-hd/v2 => ../go-eth2-wallet-hd
