package v1

//go:generate protoc --go_out=plugins=grpc:. account.proto
//go:generate protoc --go_out=plugins=grpc:. sign.proto
//go:generate protoc --go_out=plugins=grpc:. wallet.proto
