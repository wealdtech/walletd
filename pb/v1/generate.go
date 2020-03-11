package v1

//go:generate protoc --go_out=plugins=grpc:. accountmanager.proto
//go:generate protoc --go_out=plugins=grpc:. eth2.proto
//go:generate protoc --go_out=plugins=grpc:. lister.proto
//go:generate protoc --go_out=plugins=grpc:. signer.proto
//go:generate protoc --go_out=plugins=grpc:. walletmanager.proto
