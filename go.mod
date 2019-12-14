module github.com/wealdtech/walletd

go 1.13

require (
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/rpc v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/viper v1.6.1
	github.com/wealdtech/go-eth2-wallet v1.6.0
	github.com/wealdtech/go-eth2-wallet-store-filesystem v1.4.0
	github.com/wealdtech/go-eth2-wallet-store-s3 v1.4.0
	github.com/wealdtech/go-eth2-wallet-types v1.7.0
	github.com/wealdtech/go-grpcserver v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20191206224255-0243a4be9c8f
	google.golang.org/grpc v1.25.1
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace github.com/wealdtech/go-grpcserver => ../go-grpcserver
